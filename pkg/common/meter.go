package common

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	openTelemetryMeter "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"time"
)

const (
	CountSuffix    = "_count"
	LatencySuffix  = "_latency"
	ErrorSuffix    = "_error"
	ServiceNameTag = "service_name"
	ServiceEnvTag  = "service_env"
	URITag         = "uri"
	StatusCodeTag  = "status_code"
)

// Meter is a wrapper around the OpenTelemetry meter.
type Meter struct {
	appConfig *AppConfig
	otMeter   openTelemetryMeter.Meter
	env       string
}

// InitializeOpenTelemetry initializes the OpenTelemetry providers.
func InitializeOpenTelemetry(appConfig *AppConfig) *Meter {
	ctx := context.Background()

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(appConfig.OpenTelemetry.ServiceName),
		semconv.ServiceVersion(appConfig.OpenTelemetry.ServiceVersion),
		semconv.DeploymentEnvironment(GetCurrentEnvironment()),
	)

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(appConfig.OpenTelemetry.ExporterOLTPEndpoint))
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(
			time.Second*time.Duration(appConfig.OpenTelemetry.ExporterIntervalInSeconds)))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(appConfig.OpenTelemetry.ExporterOLTPEndpoint))
	if err != nil {
		panic(err)
	}

	idg := xray.NewIDGenerator()
	tracerProvider := trace.NewTracerProvider(
		trace.WithIDGenerator(idg),
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// STEP 3: Configure TextMapPropagator to extract trace IDs from incoming HTTP headers
	// This is CRITICAL - without this, trace IDs from parent requests won't be extracted
	// Composite propagator supports multiple formats:
	// - xray.Propagator{}: AWS X-Ray format (traceparent header with X-Ray format)
	// - propagation.TraceContext{}: W3C TraceContext standard (traceparent header)
	// - propagation.Baggage{}: W3C Baggage standard (baggage header)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			xray.Propagator{},              // AWS X-Ray format support
			propagation.TraceContext{},     // W3C standard traceparent header
			propagation.Baggage{},          // W3C baggage propagation
		),
	)

	return &Meter{
		appConfig: appConfig,
		otMeter:   otel.Meter(appConfig.OpenTelemetry.ServiceName),
		env:       GetCurrentEnvironment(),
	}
}

// Latency returns a latency histogram.
func (m *Meter) Latency(key string) (openTelemetryMeter.Int64Histogram, error) {
	return m.otMeter.Int64Histogram(
		key+LatencySuffix,
		openTelemetryMeter.WithUnit("ms"),
		openTelemetryMeter.WithDescription("Latency in ms"),
	)
}

// Count returns a counter.
func (m *Meter) Count(key string) (openTelemetryMeter.Int64Counter, error) {
	return m.otMeter.Int64Counter(
		key+CountSuffix,
		openTelemetryMeter.WithDescription(fmt.Sprintf("Number of %s", key)),
	)
}

// RequestResponseAddMetricAttributes returns the attributes for the request-response metrics.
func (m *Meter) RequestResponseAddMetricAttributes(c echo.Context) []openTelemetryMeter.AddOption {
	return []openTelemetryMeter.AddOption{
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   ServiceNameTag,
			Value: attribute.StringValue(m.appConfig.OpenTelemetry.ServiceName),
		}),
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   ServiceEnvTag,
			Value: attribute.StringValue(GetCurrentEnvironment()),
		}),
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   URITag,
			Value: attribute.StringValue(c.Request().URL.Path),
		}),
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   StatusCodeTag,
			Value: attribute.IntValue(c.Response().Status),
		}),
	}
}

// RequestResponseRecordMetricAttributes returns the attributes for the request-response metrics.
func (m *Meter) RequestResponseRecordMetricAttributes(c echo.Context) []openTelemetryMeter.RecordOption {
	return []openTelemetryMeter.RecordOption{
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   ServiceNameTag,
			Value: attribute.StringValue(m.appConfig.OpenTelemetry.ServiceName),
		}),
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   ServiceEnvTag,
			Value: attribute.StringValue(GetCurrentEnvironment()),
		}),
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   URITag,
			Value: attribute.StringValue(c.Request().URL.Path),
		}),
		openTelemetryMeter.WithAttributes(attribute.KeyValue{
			Key:   StatusCodeTag,
			Value: attribute.IntValue(c.Response().Status),
		}),
	}
}
