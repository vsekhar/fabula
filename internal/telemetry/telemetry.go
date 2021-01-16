package telemetry

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
)

// InitTelemetry initializes telemetry.
//
// InitTelemetry should be called from main(), and the returned function should
// be defered.
func InitTelemetry(ctx context.Context) func() {
	gpusher, err := mexporter.InstallNewPipeline(
		[]mexporter.Option{
			// mexporter.WithProjectID("TODO"), // TODO: fix this
		},
		[]push.Option{}...,
	)
	if err != nil {
		log.Fatal(err)
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	texp, err := texporter.NewExporter(texporter.WithProjectID(projectID))
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(texp))
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(gpusher.MeterProvider())
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)
	if err := host.Start(); err != nil {
		log.Fatal(err)
	}
	return func() {
		if err := texp.Shutdown(ctx); err != nil {
			log.WithError(err).Error("shutting down trace exporter")
		}
		gpusher.Stop()
	}
}
