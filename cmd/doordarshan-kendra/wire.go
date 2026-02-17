//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package cmd

import (
	"github.com/google/wire"
	"pkg/app"
	"pkg/clients"
	"pkg/common"
	"pkg/data"
	"pkg/handler"
	"pkg/server"
	"pkg/sfu"
)

// InitializeApp init echo application.
func InitializeApp(appConfig *common.AppConfig, metric *common.Meter, tenantClusterHandlerMap map[string]sfu.IClusterHandler) (*app.Application, error) {
	wire.Build(
		clients.ProviderSet,
		data.ProviderSet,
		handler.ProviderSet,
		common.ProviderSet,
		server.ProviderSet,
		app.NewApplication,
	)
	return &app.Application{}, nil
}
