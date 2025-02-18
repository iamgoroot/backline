package oauth2

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamgoroot/backline/pkg/core"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func getAuthMiddleware(logger core.Logger, pluginCfg *oAuth2Cfg) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(pluginCfg.JWTSecret),
		TokenLookup: "cookie:user-session",
		Skipper: func(c echo.Context) bool {
			url := c.Request().URL.String()
			return url == loginURL || url == loggedOutURL || strings.HasPrefix(url, callbackURL)
		},
		ErrorHandler: func(reqCtx echo.Context, err error) error {
			if err != nil {
				userSession, _ := reqCtx.Cookie("user-session")
				if userSession != nil && userSession.Value == "logged-out" {
					return reqCtx.Redirect(http.StatusTemporaryRedirect, loggedOutURL)
				}
				logger.Error(err.Error())
			}
			return reqCtx.Redirect(http.StatusTemporaryRedirect, loginURL)
		},
		SuccessHandler: func(reqCtx echo.Context) {
			user, ok := reqCtx.Get("user").(*jwt.Token)
			if !ok {
				return
			}
			sub, err := user.Claims.GetSubject()
			if err != nil {
				logger.Error(err.Error())
			} else {
				reqCtx.Set("user-email", sub)
			}
		},
	})
}
