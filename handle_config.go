package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/naoina/toml"
)

type Config struct {
	Database struct {
		User     string
		Password string
		Name     string
		Host     string
	}
	Channellogger struct {
		Token     string
		ChannelID string
	}
	Server struct {
		Address      string
		WriteTimeout time.Duration
		ReadTimeout  time.Duration
		IdleTimeout  time.Duration
	}
	Csrf struct {
		Token string
	}
	Cookies struct {
		HashKey string
	}
}

func ParseConfig() (Config, error) {
	var config Config

	filepath, err := filepath.Abs("conf.toml")
	if err != nil {
		return config, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return config, err
	}

	if err := toml.NewDecoder(file).Decode(&config); err != nil {
		return config, err
	}

	defer file.Close()

	return config, err
}
