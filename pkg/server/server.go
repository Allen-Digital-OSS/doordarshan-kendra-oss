package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/handler"
	appLog "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/log"
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	echopprof "github.com/ttys3/echo-pprof/v4"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ShutdownTimeout = 30 * time.Second
	ContextTimeout  = 5 * time.Second
)

type Server struct {
	e                *echo.Echo
	appConfig        *common.AppConfig
	appMiddleware    *AppMiddleware
	meetingV1handler handler.IMeetingV1Handler
	logger           *appLog.Logger
}

var ProviderSet = wire.NewSet(NewAppMiddleware, NewServer)

// NewServer creates a new server instance.
func NewServer(
	appConfig *common.AppConfig,
	appMiddleware *AppMiddleware,
	meetingV1Handler handler.IMeetingV1Handler) *Server {
	e := echo.New()
	echopprof.Wrap(e)
	return &Server{
		e:                e,
		appConfig:        appConfig,
		appMiddleware:    appMiddleware,
		meetingV1handler: meetingV1Handler,
	}
}

// Start starts the server.
func (s *Server) Start() {
	// STEP 5.1: Initialize zap logger with trace ID support
	// Convert config string to our LogLevel enum
	logLevel := appLog.GetLogLevelFromString(s.appConfig.Server.LogLevel)
	logger, err := appLog.NewLogger("doordarshan-kendra", logLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	s.logger = logger

	// Keep Echo's built-in logger for startup messages
	s.e.Logger.SetPrefix("[Server]")
	s.e.Logger.SetLevel(common.GetLogLevelFromString(s.appConfig.Server.LogLevel))

	// STEP 5.2: Add trace ID middleware FIRST
	// This is critical - it must run before other middleware to extract trace IDs
	s.e.Use(TraceIDMiddleware())

	// Standard Echo middleware
	s.e.Use(middleware.RequestID()) // Generates request ID
	s.e.Use(middleware.Logger())
	s.e.Use(middleware.Recover())
	s.e.Use(middleware.CORS())
	s.e.Use(WithRequestIDMiddleware()) // Stores request ID in context

	// Application middleware
	s.e.Use(s.appMiddleware.AuthNMiddleware)
	s.e.Use(s.appMiddleware.MetricsMiddleware)
	s.e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Second * 40,
	}))

	// Custom request/response logger middleware (must be after trace middleware)
	s.e.Use(LogRequestResponse)

	// Setup routes.
	s.SetupRoutes()

	// Configure server.
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.appConfig.Server.Port),
		ReadTimeout:    time.Second * time.Duration(s.appConfig.Server.ReadTimeoutInSeconds),
		WriteTimeout:   time.Second * time.Duration(s.appConfig.Server.WriteTimeoutInSeconds),
		MaxHeaderBytes: s.appConfig.Server.MaxHeaderBytes,
	}

	// Start server.
	go func() {
		s.e.Logger.Infof("Server is listening on PORT: %d", s.appConfig.Server.Port)
		err := s.e.StartServer(server)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.e.Logger.Fatalf("Error starting Server: %v", err)
			s.e.Logger.Fatal("Shutting down the server")
		}
	}()

	// Wait for interrupt meetingsignals to gracefully shut down the server with a timeout of 30 seconds.
	// Use a buffered channel to avoid missing signals as recommended for meetingsignals.Notify.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Shutdown the server post receiving quit meetingsignals.
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := s.e.Server.Shutdown(ctx); err != nil {
		s.e.Logger.Fatal(err)
	} else {
		s.e.Logger.Info("Shutting down the server")
	}
}

func (*Server) Cleanup() {
	// Cleanup logic goes here.
	log.Info("Server cleanup. Nothing to cleanup for now. Doing nothing.")
}
