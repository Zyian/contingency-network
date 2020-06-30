package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zyian/contingency-network/api"
	ssentry "github.com/Zyian/contingency-network/sentry"
	"github.com/Zyian/contingency-network/server"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

func init() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		logrus.Warn("no sentry dsn provided, sentry is essentially disabled")
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		logrus.Warn("environment is not set, defaulting to staging")
	}

	sentryOpts := sentry.ClientOptions{
		Dsn:         dsn,
		Environment: env,
	}

	hook, err := ssentry.NewHook(sentryOpts, logrus.PanicLevel, logrus.ErrorLevel, logrus.FatalLevel)
	if err != nil {
		logrus.Fatal("could not create sentry hook")
	}
	logrus.AddHook(hook)
}

func main() {
	srv := server.NewServer(api.CreateRouter())

	go srv.Start(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	logrus.Info("server running...")
	<-c
	logrus.Info("interrupt signal captured: shutting down")
	sentry.Flush(2 * time.Second)
	_ = srv.Shutdown(context.Background())
	logrus.Info("server shut down successful")
}
