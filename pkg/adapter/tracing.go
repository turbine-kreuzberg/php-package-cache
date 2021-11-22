package adapter

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// SetupTracing sets up jaeger as a global tracer for opentracing,
// a closer func is returned that can be used to flush buffers before shutdown.
func SetupTracing() (io.Closer, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, fmt.Errorf("load config from Env Vars: %v", err)
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}

	opentracing.SetGlobalTracer(tracer)

	return closer, nil
}
