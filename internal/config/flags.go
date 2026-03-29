package config

import (
	"flag"
	"fmt"
	"net"
)

type AddressWithPortFlag string

func (o *AddressWithPortFlag) String() string {
	return string(*o)
}

func (a *AddressWithPortFlag) Set(value string) error {
	if value == "" {
		value = "http://localhost:8080"
	}

	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("wrong format. Example: <addr>:<port>")
	}

	if host == "" {
        return fmt.Errorf("host cannot be empty")
    }
    if port == "" {
        return fmt.Errorf("port cannot be empty")
    }

	*a = AddressWithPortFlag(value)

	return nil
}

type Config struct {
	AppAddress AddressWithPortFlag
	BaseShortURL AddressWithPortFlag
}

func NewConfig() *Config {
	config := new(Config)
	flag.Var(&config.AppAddress, "a", "application launch address (for example, localhost:8888)")
	flag.Var(&config.BaseShortURL, "b", "the base address for the short URL")
	flag.Parse()

	return config
}