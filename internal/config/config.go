package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080"
	DefaultFileStoragePath = "storage.json"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func (c *Config) ParseFlags() {
	flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	flagSet.StringVar(&c.ServerAddress, "a", DefaultServerAddress, "application launch address (for example, localhost:8888)")
	flagSet.StringVar(&c.BaseURL, "b", DefaultBaseURL, "the base address for the short URL")
	flagSet.StringVar(&c.FileStoragePath, "f", DefaultFileStoragePath, "file storage path")
	flagSet.StringVar(&c.DatabaseDSN, "d", "", "database connection address")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Printf("flag parsing error: %v", err)
	}
}

func (c *Config) ParseEnv() {
	err := env.Parse(c)
	if err != nil {
		log.Printf("env parsing error: %v. Use flags", err)
	}
}

func (c *Config) normalize() {
	host, port, err := net.SplitHostPort(c.ServerAddress)
	if err != nil {
		c.ServerAddress = DefaultServerAddress
	}

	if host == "" {
		c.ServerAddress = fmt.Sprintf("localhost:%s", port)
	}

	if port == "" {
		c.ServerAddress = fmt.Sprintf("%s:8080", host)
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
