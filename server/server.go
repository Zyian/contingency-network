package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	hps *http.Server
}

func NewServer(r *mux.Router) *Server {
	return &Server{
		hps: &http.Server{
			Handler:      r,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	if err := s.hps.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithFields(logrus.Fields{"at": "startup", "err": err}).Error("could not start listening")
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.hps.Shutdown(ctx); err != nil {
		if err == http.ErrServerClosed {
			logrus.WithFields(logrus.Fields{"at": "shutdown", "err": err}).Errorf("trying to shutdown an already shutting down server")
			return nil
		}
		logrus.WithFields(logrus.Fields{"at": "shutdown", "err": err}).Errorf("trying to shutdown an already shutting down server")
		return err
	}
	return nil
}
