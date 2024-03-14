package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var AppConfig *Config

type Config struct {
	ENV              string        `mapstructure:"ENV"`
	TCPPort          int           `mapstructure:"TCP_PORT"`
	HTTPPort         int           `mapstructure:"HTTP_PORT"`
	DB_HOST          string        `mapstructure:"DB_HOST"`
	DB_PORT          string        `mapstructure:"DB_PORT"`
	DB_USER          string        `mapstructure:"DB_USER"`
	DB_DRIVER        string        `mapstructure:"DB_DRIVER"`
	DB_PASSWORD      string        `mapstructure:"DB_PASSWORD"`
	DB_NAME          string        `mapstructure:"DB_NAME"`
	KASPI_QR_URL     string        `mapstructure:"KASPI_QR_URL"`
	KASPI_LOGIN      string        `mapstructure:"KASPI_LOGIN"`
	KASPI_PASSWORD   string        `mapstructure:"KASPI_PASSWORD"`
	KASPI_REFUND_URL string        `mapstructure:"KASPI_REFUND_URL"`
	REFUND_TIME      time.Duration `mapstructure:"REFUND_TIME"`
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
