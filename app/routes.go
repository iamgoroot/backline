package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

const csrfEchoContextKey = "csrfEchoContextKey"

func (a *App) createServer(logger *slog.Logger, cfg *CoreCfg) *echo.Echo {
	router := echo.New()
	router.Use(middleware.Recover())

	const logLvlNone = 99

	router.Logger.SetLevel(logLvlNone)

	loggerMiddleware := slogecho.NewWithConfig(logger, slogecho.Config{
		DefaultLevel: slog.LevelDebug,
	})
	router.Use(loggerMiddleware)

	router.HTTPErrorHandler = func(err error, reqCtx echo.Context) {
		switch err.(type) {
		case core.NotFoundError:
			writeErrResponse(reqCtx, http.StatusNotFound, err)
		case core.ThirdPartyError:
			writeErrResponse(reqCtx, http.StatusBadGateway, err)
		default:
			router.DefaultHTTPErrorHandler(err, reqCtx)
		}
	}

	setupMiddlewares(router, cfg)

	return router
}

func setupMiddlewares(router *echo.Echo, cfg *CoreCfg) {
	setupCSRFMiddleware(router, cfg)
	setupCORSMiddleware(router, cfg)

	router.GET("/liveness", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
}

func setupCSRFMiddleware(router *echo.Echo, cfg *CoreCfg) {
	if !cfg.Server.CSRF.Disabled {
		return
	}

	router.Use(middleware.CSRFWithConfig(
		middleware.CSRFConfig{
			CookieSameSite: http.SameSiteStrictMode,
			CookieSecure:   !cfg.Server.CSRF.InsecureCookie,
			CookieHTTPOnly: true,
			TokenLookup:    fmt.Sprintf("header:%s,query:%s", echo.HeaderXCSRFToken, core.CSRFQueryParamName),
			ContextKey:     csrfEchoContextKey,
			CookieDomain:   cfg.Server.Host,
		}),
	)
	router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.WithValue(c.Request().Context(), core.CSRFTokenContextKey, getCSRFTokenFromEcho(c)) // put csrf token into context
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	})
}

func setupCORSMiddleware(router *echo.Echo, cfg *CoreCfg) {
	if !cfg.Server.CORS.Disabled {
		router.Use(middleware.CORSWithConfig(
			middleware.CORSConfig{
				AllowCredentials: true,
				AllowOrigins:     cfg.Server.CORS.Origins,
			}),
		)
	}
}

func writeErrResponse(c echo.Context, status int, err error) {
	respErr := c.String(status, err.Error())
	if respErr != nil {
		c.Logger().Error(respErr.Error())
	}
}
func getCSRFTokenFromEcho(c echo.Context) string {
	value := c.Get(csrfEchoContextKey)
	csrf, _ := value.(string)

	return csrf
}
