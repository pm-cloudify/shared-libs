package config_loader

import (
	"os"

	"github.com/spf13/viper"
)

// loads configuration
// set your default envs for development and production in
// .env.development and .env.production
// or use a default .env file without when no APP_ENV is set
func LoadEnv(path string) {
	if path == "" {
		path = "."
	}

	if os.Getenv("APP_ENV") != "" {
		viper.SetConfigName(".env." + os.Getenv("APP_ENV"))
	} else {
		viper.SetConfigFile(".env")
	}

	viper.AddConfigPath(path)
	viper.AutomaticEnv()
}
