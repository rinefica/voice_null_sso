package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string        `yaml:"env"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	GRPC        GRPCConfig    `yaml:"grpc"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-default:"1h"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"44044"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config file path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not exist " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config is incorrect " + err.Error())
	}
	return &cfg
}

// priority: flag > env > default
func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
