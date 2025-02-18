package oauth2

import (
	"golang.org/x/oauth2"
)

type oAuth2Cfg struct {
	Endpoint     endpointCfg `yaml:"endpoint"`
	UserInfo     userInfoCfg `yaml:"userInfo"`
	JWTSecret    string      `yaml:"jwtSecret"`
	ClientID     string      `yaml:"clientId"`
	ClientSecret string      `yaml:"clientSecret"`
	RedirectURL  string      `yaml:"redirectUrl"` // TODO: read server cfg and derive from there
	Scopes       []string    `yaml:"scopes"`
	Disabled     bool        `yaml:"disabled"`
}

type endpointCfg struct {
	AuthURL  string `yaml:"authUrl"`
	TokenURL string `yaml:"tokenUrl"`
}

type userInfoCfg struct {
	BodyDecoder       string `yaml:"bodyDecoder"`
	URL               string `yaml:"url"`
	EmailPath         string `yaml:"emailPath"`
	EmailVerifiedPath string `yaml:"emailVerifiedPath"`
}

func createOauth2Cfg(cfg *oAuth2Cfg) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.Endpoint.AuthURL,
			TokenURL: cfg.Endpoint.TokenURL,
		},
	}
}
