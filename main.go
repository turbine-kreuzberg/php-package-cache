package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ppc "github.com/turbine-kreuzberg/php-package-cache/pkg"
	"github.com/turbine-kreuzberg/php-package-cache/pkg/adapter"
	"github.com/turbine-kreuzberg/php-package-cache/pkg/middleware"
	cli "github.com/urfave/cli/v2"
)

var (
	gitHash string
	gitRef  string
)

var addr = &cli.StringFlag{Name: "addr", Value: ":8080", Usage: "Address to serve on."}
var s3_endpoint = &cli.StringFlag{Name: "s3-endpoint", Required: true, Usage: "s3 endpoint."}
var s3_access_key_file = &cli.StringFlag{Name: "s3-access-key-file", Required: true, Usage: "Path to s3 access key."}
var s3_secret_key_file = &cli.StringFlag{Name: "s3-secret-key-file", Required: true, Usage: "Path to s3 secret access key."}
var s3_ssl = &cli.BoolFlag{Name: "s3-ssl", Value: true, Usage: "s3 uses SSL."}
var s3_location = &cli.StringFlag{Name: "s3-location", Value: "us-east-1", Usage: "s3 bucket location."}
var s3_bucket = &cli.StringFlag{Name: "s3-bucket", Required: true, Usage: "s3 bucket name."}
var upstream_endpoint = &cli.StringFlag{Name: "upstream-endpoint", Value: "", Usage: "Notify service endpoint."}

func main() {
	app := &cli.App{
		Name:   "php-package-cache",
		Usage:  "Local composer package cache.",
		Action: run,
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Local composer package cache.",
				Flags: []cli.Flag{
					addr,
					s3_endpoint,
					s3_access_key_file,
					s3_secret_key_file,
					s3_ssl,
					s3_location,
					s3_bucket,
					upstream_endpoint,
				},
				Action: run,
			},
			{
				Name:  "version",
				Usage: "Show the version",
				Action: func(c *cli.Context) error {
					_, err := os.Stdout.WriteString(fmt.Sprintf("version: %s\ngit commit: %s", gitRef, gitHash))
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	log.Printf("version: %v", gitRef)
	log.Printf("git commit: %v", gitHash)

	log.Println("set up metrics")

	middleware.InitMetrics(gitHash, gitRef)

	log.Println("set up tracing")

	tracer, err := adapter.SetupTracing()
	if err != nil {
		return fmt.Errorf("setup tracing: %v", err)
	}
	defer tracer.Close()

	log.Println("set up storage")

	storage, err := adapter.SetupObjectStorage(
		c.String(s3_endpoint.Name),
		c.String(s3_access_key_file.Name),
		c.String(s3_secret_key_file.Name),
		c.Bool(s3_ssl.Name),
		c.String(s3_location.Name),
		c.String(s3_bucket.Name))
	if err != nil {
		return fmt.Errorf("setup minio s3 client: %v", err)
	}

	log.Println("set up service")

	rand.Seed(time.Now().UTC().UnixNano())

	svc := ppc.NewService(storage)
	svc = middleware.RequestID(rand.Int63, svc)
	svc = middleware.Logging(svc)
	svc = middleware.InitTraceContext(svc)
	svc = middleware.InstrumentHttpHandler(svc)
	svc = middleware.Timeout(10*time.Minute, svc)

	err = serverHttp(svc, c.String(addr.Name))
	if err != nil {
		return fmt.Errorf("run http server: %v", err)
	}

	return nil
}

func serverHttp(svc http.Handler, addr string) error {
	svcServer := adapter.NewHttpServer(svc, addr)

	log.Println("starting server")

	go adapter.MustListenAndServe(svcServer)

	log.Println("running")

	awaitShutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := svcServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("shutdown service server: %v", err)
	}

	log.Println("shutdown complete")

	return nil
}

func awaitShutdown() {
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
