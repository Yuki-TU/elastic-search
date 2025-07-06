package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port             string `env:"PORT" envDefault:"8080"`
	Environment      string `env:"ENVIRONMENT" envDefault:"development"`
	ElasticsearchURL string `env:"ELASTICSEARCH_URL" envDefault:"http://localhost:9200"`
}

func NewConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return cfg
}
