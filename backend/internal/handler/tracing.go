package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer sets up the OTel TracerProvider with a gRPC OTLP exporter.
// Returns a shutdown function that should be called on graceful shutdown.
func InitTracer(ctx context.Context, endpoint string) func(context.Context) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("WARN: failed to connect to OTel collector at %s: %v (tracing disabled)", endpoint, err)
		return func(context.Context) {}
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Printf("WARN: failed to create OTLP trace exporter: %v (tracing disabled)", err)
		return func(context.Context) {}
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("backend"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		log.Printf("WARN: failed to create OTel resource: %v", err)
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	log.Printf("OTel tracing initialized, exporting to %s", endpoint)

	return func(ctx context.Context) {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("WARN: failed to shutdown TracerProvider: %v", err)
		}
	}
}

// OTelMiddleware creates a trace span for every HTTP request using otelhttp.
// It uses the chi route pattern as the span name for low cardinality.
func OTelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewMiddleware("backend",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}
			return r.Method + " " + routePattern
		}),
	)(next)
}
