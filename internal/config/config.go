package config

import "time"

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
