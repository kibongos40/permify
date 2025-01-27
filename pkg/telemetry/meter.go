package telemetry

import (
	"context"
	"os"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"go.opentelemetry.io/contrib/instrumentation/host"
	orn "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	omt "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/Permify/permify/internal"
)

// NewMeter - Creates new meter
func NewMeter(exporter metric.Exporter, interval time.Duration) func(context.Context) error {
	hostName, err := os.Hostname()
	if err != nil {
		return func(context.Context) error { return nil }
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(
			exporter,
			metric.WithInterval(interval),
		)),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("permify"),
			attribute.String("id", internal.Identifier),
			attribute.String("project.id", internal.Identifier),
			attribute.String("version", internal.Version),
			attribute.String("host_name", hostName),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		)),
	)

	if err = orn.Start(
		orn.WithMinimumReadMemStatsInterval(time.Second),
		orn.WithMeterProvider(mp),
	); err != nil {
		return func(context.Context) error { return nil }
	}

	if err = host.Start(host.WithMeterProvider(mp)); err != nil {
		return func(context.Context) error { return nil }
	}

	otel.SetMeterProvider(mp)

	return mp.Shutdown
}

// NewNoopMeter - Creates new noop meter
func NewNoopMeter() omt.Meter {
	mp := noop.MeterProvider{}
	return mp.Meter("permify")
}
