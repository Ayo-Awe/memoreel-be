package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/ayo-awe/memoreel-be/util"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

var config applicationConfiguration

type ConfigType string

const (
	Test ConfigType = "test"
	Prod ConfigType = "production"
)

type applicationConfiguration struct {
	TestConfiguration       Configuration `env:", prefix=TEST_"`
	ProductionConfiguration Configuration
}

type Configuration struct {
	Database DatabaseConfiguration
	Server   ServerConfiguration
}

type DatabaseConfiguration struct {
	Username string `env:"DB_USERNAME, default=postgres"`
	Password string `env:"DB_PASSWORD, default=postgres"`
	Host     string `env:"DB_HOST, default=localhost"`
	Database string `env:"DB_DATABASE, default=memoreel"`
	Port     int    `env:"DB_PORT, default=5432"`
	SSLMode  string `env:"DB_SSL_MODE, default=disable"`
}

type ServerConfiguration struct {
	Port int `env:"PORT, default=8080"`
}

func (d DatabaseConfiguration) BuildDSN() string {
	dsnFormat := "postgres://%s@%s/%s?sslmode=%s"

	auth := fmt.Sprintf("%s:%s", d.Username, d.Password)
	address := fmt.Sprintf("%s:%d", d.Host, d.Port)

	return fmt.Sprintf(dsnFormat, auth, address, d.Database, d.SSLMode)
}

func Get(configType ConfigType) Configuration {
	if configType == Test {
		return config.TestConfiguration
	}

	return config.ProductionConfiguration
}

// This must be called before get for the config to function properly
func LoadConfig() error {
	projectRoot, err := util.GetProjectRoot()
	if err != nil {
		return err
	}

	envPath := path.Join(projectRoot, ".env")

	// load envs from .env file if it exists
	err = godotenv.Load(envPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err = envconfig.Process(context.Background(), &config)
	if err != nil {
		return err
	}

	return nil
}
