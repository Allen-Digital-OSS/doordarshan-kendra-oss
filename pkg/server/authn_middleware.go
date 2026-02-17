package server

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
)

const (
	AuthorizationHeaderKey = "Authorization"
	UnauthenticatedUser    = "unauthenticated_user"
)

var URLPathToSkipAuth = []string{"/health", "/"}

// AuthNMiddleware From Access token get claims and validate if all required claims are there or not and if not return 401.
func (am *AppMiddleware) AuthNMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// Skip user token parsing for health check.
		if slices.Contains(URLPathToSkipAuth, ctx.Path()) {
			return next(ctx)
		}

		// Default user claims.
		defaultUserClaims := &common.UserInfoFromJWTClaims{
			UserID:      UnauthenticatedUser,
			PersonaType: UnauthenticatedUser,
			TenantID:    UnauthenticatedUser,
			ExternalID:  UnauthenticatedUser,
		}

		// Skip user token parsing if flag is set.
		if am.appConfig.Server.SkipUserTokenAuthentication {
			ctx.Set(constant.UserInfoFromJWTClaimsKey, defaultUserClaims)
			return next(ctx)
		}

		// Get logger instance once for use throughout the function
		logger := GetLoggerFromEchoContext(ctx)

		// Read authorization header and extract JWT token from the request.
		authHeader := ctx.Request().Header.Get(AuthorizationHeaderKey)
		accessToken, err := common.ExtractAccessTokenFromAuthorizationHeader(authHeader)
		if err != nil {
			if logger != nil {
				logger.Errorf(ctx.Request().Context(), "[AuthNMiddleware] Error while extracting access token from auth header err: %v", err)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		// Extract claims from access token.
		if logger != nil {
			logger.Infof(ctx.Request().Context(), "[AuthNMiddleware] Extracting claims from access token")
		}
		userInfoFromJWTClaims, err := common.ExtractClaimsFromToken(accessToken, am.appConfig.Server.JWTSecret)
		if err != nil {
			if logger != nil {
				logger.Errorf(ctx.Request().Context(), "[AuthNMiddleware] Error while parsing Access Token and extracting claims err: %v", err)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		ctx.Set(constant.UserInfoFromJWTClaimsKey, userInfoFromJWTClaims)
		return next(ctx)
	}
}
