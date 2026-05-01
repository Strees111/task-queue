package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Address   string        `yaml:"address" env:"ADDRESS" env-default:"localhost:2342"`
	LogLevel  string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Timeout   time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s"`
	Workers   int           `yaml:"workers" env:"WORKERS" env-default:"10"`
	QueueSize int           `yaml:"queue_size" env:"QUEUE_SIZE" env-default:"64"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("read config failed: %v", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("read env failed: %v", err)
		}
	}
	return cfg
}
