package main

import "github.com/caarlos0/env/v9"

type Config struct {
	Database struct {
		Host     string `env:"PSQL_HOST,required"`
		Port     int    `env:"PSQL_PORT,required"`
		User     string `env:"PSQL_USER,required"`
		Password string `env:"PSQL_PASSWORD,required"`
		Name     string `env:"PSQL_DATABASE,required"`
	}
}

func NewConfigFromEnv() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
