package main

import (
	"os"
	"strings"

	"github.com/jinzhu/configor"
)

type Config struct {
	Server struct {
		Address  string `toml:"address" required:"true"`
		Username string `toml:"username" required:"true"`
		Password string `toml:"password" required:"true"`
	} `toml:"server"`
	Session struct {
		Path string `toml:"path"`
	} `toml:"session"`
}

func NewConfig(path string) (*Config, error) {
	config := &Config{}
	err := configor.Load(config, path)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(config.Session.Path, "~/") {
		config.Session.Path = os.Getenv("HOME") + "/" +
			strings.TrimPrefix(config.Session.Path, "~/")
	}

	return config, nil
}
