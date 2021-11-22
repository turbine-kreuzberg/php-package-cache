package adapter

import (
	"log"
	"net/http"
	"time"
)

func NewHttpServer(h http.Handler, addr string) *http.Server {
	httpServer := &http.Server{
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}
	httpServer.Addr = addr
	httpServer.Handler = h

	return httpServer
}

func MustListenAndServe(srv *http.Server) {
	log.Printf("starting http server on %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
