package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port              string `envconfig:"PORT"                default:":4000"`
	AppEnv            string `envconfig:"APP_ENV"             default:"development"`
	DatabaseURL       string `envconfig:"DATABASE_URL"`
	DatabaseURLDev    string `envconfig:"DATABASE_URL_DEV"`
	SessionCookieName string `envconfig:"SESSION_COOKIE_NAME" default:"session"`
	SessionSecret     string `envconfig:"SESSION_SECRET"      default:"VERYSECRETKEY"`
	SkipMigrate       bool   `envconfig:"SKIP_MIGRATE"        default:"false"`
}

func (c *Config) ActiveDatabaseURL() string {
	if c.AppEnv == "development" {
		if c.DatabaseURLDev == "" {
			fmt.Fprintln(
				os.Stderr,
				"APP_ENV=development mas DATABASE_URL_DEV não definida",
			)
			os.Exit(1)
		}
		return c.DatabaseURLDev
	}
	if c.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL não definida")
		os.Exit(1)
	}
	return c.DatabaseURL
}

func loadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
