//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package cmd

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/app"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/clients"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/data"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/handler"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/server"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/sfu"
	"github.com/google/wire"
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
