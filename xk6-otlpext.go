package otlpext

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	modules.Register("k6/x/otlpext", new(OTLPExt))
}

type OTLPExt struct {
	initialized bool
	tp          *sdktrace.TracerProvider
	mtx         sync.Mutex
	counter     uint64

	OtlpTracesEndpoint string
	ServiceName        string
	Counter            uint64
}

// types
const DefaultOtlpTracesEndpoint = "http://localhost:4317"
const TracerName = "bench"

func (o *OTLPExt) InitProvider() error {
	if o.initialized {
		return nil
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.initialized {
		return nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(o.ServiceName),
			semconv.ServiceVersionKey.String("0.0.1"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var otlpTracesEndpoint string = DefaultOtlpTracesEndpoint

	if strings.HasPrefix(otlpTracesEndpoint, "http://") {
		otlpTracesEndpoint = otlpTracesEndpoint[7:]
	} else if strings.HasPrefix(otlpTracesEndpoint, "https://") {
		otlpTracesEndpoint = otlpTracesEndpoint[8:]
	} else {
		log.Fatal("endpoint must start with either http:// or https://")
	}

	conn, err := grpc.DialContext(ctx,
		otlpTracesEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	o.tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(o.tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Info("Tracer initialized")
	o.initialized = true
	return nil
}

func (o *OTLPExt) Initialize(otlpTracesEndpoint string, serviceName string) {
	o.OtlpTracesEndpoint = otlpTracesEndpoint
	o.ServiceName = serviceName
	err := o.InitProvider()
	if err != nil {
		log.Errorf("err: %w", err)
	}
}

func (o *OTLPExt) Shutdown() {
	if o.tp != nil {
		o.tp.Shutdown(context.Background())
	}
}

func (o *OTLPExt) SendTrace(spanName string) string {
	var span trace.Span
	ctx := context.Background()
	log.Debugf("Sending a span '%s' as a service '%s'", spanName, TracerName)

	ctx, span = otel.Tracer(TracerName).Start(ctx, spanName)
	atomic.AddUint64(&o.counter, 1)
	o.Counter = atomic.LoadUint64(&o.counter)
	defer span.End()

	log.Debugf("TraceID: %s, SpanID: %s", span.SpanContext().TraceID(), span.SpanContext().SpanID())
	return span.SpanContext().TraceID().String()
}
