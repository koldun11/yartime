package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koldun11/yartime/server/config"
	"go.uber.org/zap"
	"net/http"
)

// Server управляет HTTP-сервером
type Server struct {
	server *http.Server
	logger *zap.Logger
}

// NewServer создаёт новый Server
func NewServer(conf *config.AppConfig, logger *zap.Logger, r *gin.Engine) *Server {
	addr := fmt.Sprintf(":%s", conf.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &Server{
		server: srv,
		logger: logger,
	}
}

// Start запускает сервер
func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.String("addr", s.server.Addr))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server failed to start", zap.Error(err))
		}
	}()
	return nil
}

// Stop останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping server")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown failed", zap.Error(err))
		return err
	}
	return nil
}
