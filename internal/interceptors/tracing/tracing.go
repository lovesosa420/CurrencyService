package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"time"
)

func NewTracer() (opentracing.Tracer, io.Closer, error) {
	cfg := config.Configuration{
		ServiceName: "grpc-tracer",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "localhost:6831",
		},
	}

	return cfg.NewTracer()
}
