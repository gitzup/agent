package logger

import (
	"context"
	log "github.com/sirupsen/logrus"
	golog "log"
	"os"
	"time"
)

var root *log.Entry

// Initialize the logging infrastructure.
func InitLogger(version string, caller bool, logLevel string, logFormat string) {
	log.SetOutput(os.Stdout)
	log.SetReportCaller(caller)

	// apply log level
	switch logLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		golog.Fatalf("invalid loglevel provided: %s\n", logLevel)
	}

	// apply appropriate log formatter, according to the logFormat flag
	switch logFormat {
	case "auto":
		// no-op, auto-detected by logrus
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: false,
			PrettyPrint:      false,
			TimestampFormat:  time.RFC3339,
		})
	case "plain":
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp:       false,
			DisableColors:          true,
			DisableLevelTruncation: true,
			DisableSorting:         false,
			ForceColors:            false,
			FullTimestamp:          true,
			TimestampFormat:        time.RFC3339,
		})
	case "pretty":
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp:       false,
			DisableColors:          false,
			DisableLevelTruncation: false,
			DisableSorting:         false,
			ForceColors:            true,
			FullTimestamp:          true,
			TimestampFormat:        time.RFC3339,
		})
	default:
		golog.Fatalf("invalid logformat provided: %s\n", logFormat)
	}

	// create the root logger
	root = log.WithFields(log.Fields{
		"version": version,
	})

	// redirect Golang standard log package output to logrus
	golog.SetFlags(0)
	golog.SetOutput(Logger().Writer())
}

func Logger() *log.Entry {
	if root == nil {
		panic("logger has not been set!")
	}
	return root
}

func From(ctx context.Context) *log.Entry {
	var logger = Logger()
	if requestId, ok := ctx.Value("request").(string); ok {
		logger = logger.WithField("request", requestId)
	}
	if resourceName, ok := ctx.Value("resource").(string); ok {
		logger = logger.WithField("resource", resourceName)
	}
	if containerName, ok := ctx.Value("container").(string); ok {
		logger = logger.WithField("container", containerName)
	}
	return logger
}
