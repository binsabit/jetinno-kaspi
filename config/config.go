package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

var AppConfig *Config

type Config struct {
	Port                string `mapstructure:"SERVER_PORT"`
	KASPI_SERVICE_ID    string `mapstructure:"KASPI_SERVICE_ID"`
	KASPI_PAYMENT_URL   string `mapstructure:"KASPI_PAYMENT_URL"`
	KASPI_REDIRECT_URL  string `mapstructure:"KASPI_REDIRECT_URL"`
	KASPI_REFERRED_HOST string `mapstructure:"KASPI_REFERRED_HOST"`
	DB_HOST             string `mapstructure:"DB_HOST"`
	DB_PORT             string `mapstructure:"DB_PORT"`
	DB_USER             string `mapstructure:"DB_USER"`
	DB_DRIVER           string `mapstructure:"DB_DRIVER"`
	DB_PASSWORD         string `mapstructure:"DB_PASSWORD"`
	DB_NAME             string `mapstructure:"DB_NAME"`
}

func LoadConfig() (*Config, error) {
	err := os.Setenv("TZ", "Asia/Almaty")
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("/app/docker")
	viper.AddConfigPath("/docker/")

	cfg := &Config{}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("viper could not read config file:%v", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	AppConfig = cfg

	return cfg, nil
}
