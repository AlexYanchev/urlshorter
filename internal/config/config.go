package config

import (
	"flag"
	"log"
	"net"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "application launch address (for example, localhost:8888)")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "the base address for the short URL")

	flag.Parse()
}

func (c *Config) ParseEnv() {
	err := env.Parse(c)
    if err != nil {
        log.Printf("env parsing error: %v. Use flags", err)
    }
}

func (c *Config) normalize() {
	_, _, err := net.SplitHostPort(c.ServerAddress)
	if err != nil {
		c.ServerAddress = "localhost:8080"
	}

	if c.BaseURL == "" {
		c.BaseURL = "http://" + c.ServerAddress
	}

	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		c.BaseURL = "http://" + c.BaseURL
	} 
}

func NewConfig() *Config {
	cfg := &Config{}

	cfg.ParseFlags()
	cfg.ParseEnv()

	cfg.normalize()

	return cfg
}