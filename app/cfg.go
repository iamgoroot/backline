package app

import (
	"time"
)

type CoreCfg struct {
	Logger LoggerCfg `yaml:"logger"`
	Server ServerCfg `yaml:"server"`
}

type CSRF struct {
	Disabled       bool `yaml:"disabled"`
	InsecureCookie bool `yaml:"insecureCookie"`
}

type CORS struct {
	Origins  []string `yaml:"origins"`
	Disabled bool     `yaml:"disabled"`
}

type LoggerCfg struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	AddSource bool   `yaml:"addSource"`
}

type ServerCfg struct {
	Address  string `yaml:"address"`
	Host     string `yaml:"host"`
	HTTPS    HTTPS  `yaml:"https"`
	CORS     CORS   `yaml:"cors"`
	Port     int    `yaml:"port"`
	CSRF     CSRF   `yaml:"csrf"`
	Disabled bool   `yaml:"disabled"`
}

type HTTPS struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	Disabled bool   `yaml:"disabled"`
}

type Scanner struct {
	Period time.Duration `yaml:"period"`
}

func getDefaultConfig() *CoreCfg {
	const defaultHTTPPort = 8080

	return &CoreCfg{
		Server: ServerCfg{Port: defaultHTTPPort},
	}
}
