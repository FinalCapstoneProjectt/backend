package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port            string `mapstructure:"PORT"`
	DBHost          string `mapstructure:"DB_HOST"`
	DBPort          string `mapstructure:"DB_PORT"`
	DBUser          string `mapstructure:"DB_USER"`
	DBPassword      string `mapstructure:"DB_PASSWORD"`
	DBName          string `mapstructure:"DB_NAME"`
	DBSSLMode       string `mapstructure:"DB_SSLMODE"`
	JWTSecret       string `mapstructure:"JWT_SECRET"`
	Environment     string `mapstructure:"ENVIRONMENT"`
	AIServiceURL    string `mapstructure:"AI_SERVICE_URL"`
	AIServiceAPIKey string `mapstructure:"AI_SERVICE_API_KEY"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigFile(".env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	err = viper.Unmarshal(&config)
	return
}
