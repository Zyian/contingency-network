package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Zyian/contingency-network/server"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	hook, err := NewHook(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Environment: os.Getenv("ENVIRONMENT"),
		Debug:       true,
	}, logrus.PanicLevel, logrus.ErrorLevel, logrus.FatalLevel)
	if err != nil {
		logrus.Fatal("could not create sentry hook")
	}
	log.AddHook(hook)

	r := mux.NewRouter()
	srv := server.NewServer(r)

	ctx := context.Background()
	go srv.Start(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	logrus.Info("server running...")
	<-c
	logrus.Info("interrupt signal captured: shutting down")
	_ = srv.Shutdown(ctx)
}

// port in sentry <-> logrus hook implementation from: https://github.com/onrik/logrus/blob/master/sentry/sentry.go
type Hook struct {
	client      *sentry.Client
	levels      []logrus.Level
	tags        map[string]string
	release     string
	environment string
	prefix      string
}

var (
	levelsMap = map[logrus.Level]sentry.Level{
		logrus.PanicLevel: sentry.LevelFatal,
		logrus.FatalLevel: sentry.LevelFatal,
		logrus.ErrorLevel: sentry.LevelError,
		logrus.WarnLevel:  sentry.LevelWarning,
		logrus.InfoLevel:  sentry.LevelInfo,
		logrus.DebugLevel: sentry.LevelDebug,
		logrus.TraceLevel: sentry.LevelDebug,
	}
)

func (h *Hook) Levels() []logrus.Level {
	return h.levels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	exceptions := []sentry.Exception{}

	if err, ok := entry.Data[logrus.ErrorKey].(error); ok && err != nil {
		stacktrace := sentry.ExtractStacktrace(err)
		if stacktrace == nil {
			stacktrace = sentry.NewStacktrace()
		}

		exceptions = append(exceptions, sentry.Exception{
			Type:       entry.Message,
			Value:      err.Error(),
			Stacktrace: stacktrace,
		})
	}

	event := sentry.Event{
		Level:       levelsMap[entry.Level],
		Message:     h.prefix + entry.Message,
		Extra:       map[string]interface{}(entry.Data),
		Tags:        h.tags,
		Environment: h.environment,
		Release:     h.release,
		Exception:   exceptions,
	}
	hub := sentry.CurrentHub()
	h.client.CaptureEvent(&event, nil, hub.Scope())

	return nil
}

func (h *Hook) SetPrefix(prefix string) {
	h.prefix = prefix
}

func (h *Hook) AddTag(key, value string) {
	h.tags[key] = value
}

func (h *Hook) SetRelease(release string) {
	h.release = release
}

func (h *Hook) SetEnvironment(environment string) {
	h.environment = environment
}

func NewHook(options sentry.ClientOptions, levels ...logrus.Level) (*Hook, error) {
	client, err := sentry.NewClient(options)
	if err != nil {
		return nil, err
	}

	h := Hook{
		client: client,
		levels: levels,
		tags:   map[string]string{},
	}

	if len(h.levels) == 0 {
		h.levels = logrus.AllLevels
	}

	return &h, nil
}
