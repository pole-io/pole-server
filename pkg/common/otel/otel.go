package otel

import (
	"context"
	"errors"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type OtelShutdown func(context.Context) error

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, conf *Config) (shutdown OtelShutdown, err error) {
	var shutdownFuncs []OtelShutdown

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	res, err := resource.New(context.Background(), resource.WithAttributes())
	if err != nil {
		return nil, err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx, res, conf)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, res, conf)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, res *resource.Resource, conf *Config) (*trace.TracerProvider, error) {
	traceExporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(conf.Endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(conf.Timeout),
		otlptracegrpc.WithReconnectionPeriod(conf.ReconnectionPeriod),
		otlptracegrpc.WithCompressor(conf.Compressor),
	))
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	return tracerProvider, nil
}

func Meter() otelmetric.Meter {
	meter := otel.GetMeterProvider().Meter("pole-server")
	if meter == nil {
		panic("meter provider is not initialized")
	}
	return meter
}

func newMeterProvider(ctx context.Context, res *resource.Resource, conf *Config) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(conf.Endpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithTimeout(conf.Timeout),
		otlpmetricgrpc.WithReconnectionPeriod(conf.ReconnectionPeriod),
		otlpmetricgrpc.WithCompressor(conf.Compressor),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(conf.PushInterval))),
		metric.WithReader(metric.NewManualReader(metric.WithProducer(runtime.NewProducer()))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	if err := runtime.Start(); err != nil {
		return nil, err
	}
	return meterProvider, nil
}
