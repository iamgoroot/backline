package oauth2

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
)

const (
	configPath    = "$.oauth2"
	loginURL      = "/oauth2/login"
	logoutURL     = "/oauth2/logout"
	loggedOutURL  = "/oauth2/logged-out"
	callbackURL   = "/oauth2/callback"
	logoutHTML    = `<a id="oauth2-logout-link" href="` + logoutURL + `">Logout</a>`
	loggedOutHTML = `You have been logged out. <a href="` + loginURL + `">Click here to login again</a>`
)

type Plugin struct {
	core.HeaderPlugin
}

func (Plugin) Setup(_ context.Context, deps core.Dependencies) error {
	pluginCfg := &oAuth2Cfg{}
	if err := deps.CfgReader().ReadAt(configPath, pluginCfg); err != nil {
		return err
	}

	if pluginCfg.Disabled {
		deps.Logger().Info("oauth2 plugin disabled")
		return nil
	}

	oauth2Cfg := createOauth2Cfg(pluginCfg)

	deps.Router().Use(getAuthMiddleware(deps.Logger(), pluginCfg))

	handler := oauth2Handlers{oauth2: oauth2Cfg, pluginCfg: pluginCfg, kvStore: deps.StoreKV()}
	deps.Router().GET(loginURL, handler.login)
	deps.Router().GET(logoutURL, handler.logout)
	deps.Router().GET(callbackURL, handler.callback)
	deps.Router().GET(loggedOutURL, handler.loggedOut)

	return nil
}
