package main

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	Server struct {
		Address  string `toml:"address" required:"true"`
		Username string `toml:"username" required:"true"`
		Password string `toml:"password" required:"true"`
	} `toml:"server"`
}

func NewConfig(path string) (*Config, error) {
	config := &Config{}
	err := configor.Load(config, path)
	if err != nil {
		return nil, err
	}

	return config, nil
}
