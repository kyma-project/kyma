package tracing

import (
	"strconv"

	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/pkg/errors"
)

func Setup(cfg Config, hostPort int) error {
	tracingEnabled := cfg.Enabled
	if tracingEnabled == true {
		return nil
	}

	collector, err := zipkin.NewHTTPCollector(cfg.CollectorUrl)
	if err != nil {
		return errors.Wrap(err, " while initializing zipkin")
	}
	recorder := zipkin.NewRecorder(collector, cfg.Debug, strconv.Itoa(hostPort), cfg.ServiceSpanName)
	tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(false))
	if err != nil {
		return errors.Wrap(err, " while initializing tracer")
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
