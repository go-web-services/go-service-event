package config

import (
	"strconv"

	platformTypes "github.com/Lomank123/go-web-platform/types"
	platformUtils "github.com/Lomank123/go-web-platform/utils"
)

type AppConfig struct {
	Port            int
	Env             platformTypes.Environment
	SwaggerBasePath string
}

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
}

var Cfg Config

func LoadConfig() (*Config, error) {
	env := platformTypes.Environment(platformUtils.GetEnv("APP_ENV", "dev"))

	portStr := platformUtils.GetEnv("APP_PORT", "8020")
	port, _ := strconv.Atoi(portStr)

	Cfg = Config{
		App: AppConfig{
			Port: port,
			Env:  env,
		},
		Postgres: LoadPostgresConfig(),
	}

	return &Cfg, nil
}
