package global_logger

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
	"os"
)

type Logger interface {
	Trace(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warning(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
}

type GlobalLogger struct {
	logr.Logger
}

func (l *GlobalLogger) Trace(msg string, keysAndValues ...interface{}) {
	l.V(9).Info(msg, keysAndValues...)
}

func (l *GlobalLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.V(7).Info(msg, keysAndValues...)
}

// Info prints messages about the general state of the API or SDK.
// This should usually be less then 5 messages a minute.
func (l *GlobalLogger) Info(msg string, keysAndValues ...interface{}) {
	l.V(3).Info(msg, keysAndValues...)
}

func (l *GlobalLogger) Warning(msg string, keysAndValues ...interface{}) {
	l.V(1).Info(msg, keysAndValues...)
}

type LoggingApp struct {
	Ctx      context.Context
	Logger   Logger
	shutdown func(context.Context) error
}

func (a *LoggingApp) Close() error {
	return a.shutdown(a.Ctx)
}
func InitLogging(v int) *LoggingApp {
	ctx := context.Background()
	logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))

	stdr.SetVerbosity(v)
	shutdown, err := installExportPipeline(ctx, logger)
	if err != nil {
		logger.Error(err, "failed to install export pipeline")
		log.Fatal(err)
	}
	return &LoggingApp{
		Ctx:      ctx,
		Logger:   &GlobalLogger{logger},
		shutdown: shutdown,
	}

}

func installExportPipeline(ctx context.Context, logger logr.Logger) (func(context.Context) error, error) {

	otel.SetLogger(logger)
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating stdout exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource()),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetLogger(logger)
	return tracerProvider.Shutdown, nil
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("http_proxy"),
			semconv.ServiceVersionKey.String("v0.0.0"),
			attribute.String("environment", os.Getenv("ENVIRONMENT"))),
	)

	return r
}
