package config

import (
	"flag"
	"net"
	"strings"
)

type AddressWithPortFlag string

func (o *AddressWithPortFlag) String() string {
	return string(*o)
}

func (a *AddressWithPortFlag) Set(value string) error {
	if value == "" {
		value = "localhost:8080"
	}

	_, _, err := net.SplitHostPort(value)
	if err != nil {
		value = "localhost:8080"
	}

	*a = AddressWithPortFlag(value)

	return nil
}

type Config struct {
	AppAddress AddressWithPortFlag
	BaseShortURL string
}

func NewConfig() *Config {
	config := &Config{
		AppAddress:   "localhost:8080",
		BaseShortURL: "http://localhost:8080",
	}
	flag.Var(&config.AppAddress, "a", "application launch address (for example, localhost:8888)")
	flag.StringVar(&config.BaseShortURL, "b", "", "the base address for the short URL")
	flag.Parse()

	if config.BaseShortURL == "" {
		config.BaseShortURL = "http://" + config.AppAddress.String()
	}

	if !strings.HasPrefix(config.BaseShortURL, "http://") && !strings.HasPrefix(config.BaseShortURL, "https://") {
		config.BaseShortURL = "http://" + config.BaseShortURL
	} 
	

	return config
}