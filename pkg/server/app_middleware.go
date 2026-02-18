package server

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
)

// AppMiddleware is a middleware manager.
type AppMiddleware struct {
	appConfig *common.AppConfig
	meter     *common.Meter
	origins   []string
}

// NewAppMiddleware creates a new middleware instance.
func NewAppMiddleware(appConfig *common.AppConfig, meter *common.Meter) *AppMiddleware {
	return &AppMiddleware{
		appConfig: appConfig,
		meter:     meter,
		origins:   appConfig.Server.CORSOrigins,
	}
}
