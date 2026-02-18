package app

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/server"
)

// Application is a container for all the core components of the application.
type Application struct {
	appConfig *common.AppConfig
	meter     *common.Meter
	Server    *server.Server
}

// NewApplication is a constructor for Application.
func NewApplication(
	appConfig *common.AppConfig,
	meter *common.Meter,
	serverInstance *server.Server) *Application {
	return &Application{
		appConfig: appConfig,
		meter:     meter,
		Server:    serverInstance,
	}
}

// Start starts the application.
func (a *Application) Start() {
	// Defer the cleanup function.
	defer a.Cleanup()
	// Start the server.
	a.Server.Start()
}

// Cleanup is a destructor for Application.
func (a *Application) Cleanup() {
	a.Server.Cleanup()
}
