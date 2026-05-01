package config

import (
	"fmt"

	platformUtils "github.com/go-web-services/go-web-platform/utils"
)

type PostgresConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func LoadPostgresConfig() PostgresConfig {
	return PostgresConfig{
		User:     platformUtils.GetEnv("POSTGRES_USER", "go-service-event"),
		Password: platformUtils.GetEnv("POSTGRES_PASSWORD", "go-service-event-password"),
		Host:     platformUtils.GetEnv("POSTGRES_HOST", "host.docker.internal"),
		Port:     platformUtils.GetEnv("POSTGRES_PORT", "5437"),
		DBName:   platformUtils.GetEnv("POSTGRES_DB", "go-service-event-db"),
		SSLMode:  platformUtils.GetEnv("POSTGRES_SSL_MODE", "disable"),
	}
}
