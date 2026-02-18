package handler

import (
	"errors"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	appLog "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/log"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/sfu"
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"net/http"
)

const ()

// NewLoggerProvider creates a logger instance for use in Wire dependency injection
func NewLoggerProvider(appConfig *common.AppConfig) (*appLog.Logger, error) {
	logLevel := appLog.GetLogLevelFromString(appConfig.Server.LogLevel)
	return appLog.NewLogger("doordarshan-kendra", logLevel)
}

var ProviderSet = wire.NewSet(NewMeetingV1Handler, NewLoggerProvider, sfu.NewBaseClusterHandler)

// HealthProbe checks the health status of the service
// @Summary      Health check endpoint
// @Description  Returns OK if the service is healthy and running
// @Tags         Health
// @Accept       json
// @Produce      text/plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func HealthProbe(ec echo.Context) error {
	return ec.String(http.StatusOK, "OK")
}

func HandleCommonError(err error) error {
	if errors.Is(err, echo.ErrUnsupportedMediaType) {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "only application/json content type is supported")
	}
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}
