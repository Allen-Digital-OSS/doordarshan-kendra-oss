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
	GoPool    *common.GoRoutinePool
}

// NewApplication is a constructor for Application.
func NewApplication(
	appConfig *common.AppConfig,
	meter *common.Meter,
	serverInstance *server.Server,
	pool *common.GoRoutinePool) *Application {
	return &Application{
		appConfig: appConfig,
		meter:     meter,
		Server:    serverInstance,
		GoPool:    pool,
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
	a.GoPool.Release()
}
