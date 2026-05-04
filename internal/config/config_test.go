package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_EnvPriorityOverFlag(t *testing.T) {
	origServerAddress := os.Getenv("SERVER_ADDRESS")
	origBaseURL := os.Getenv("BASE_URL")

	os.Setenv("SERVER_ADDRESS", ":8080")
	os.Setenv("BASE_URL", "http://ex.ru")

	defer func() {
		os.Setenv("SERVER_ADDRESS", origServerAddress)
		os.Setenv("BASE_URL", origBaseURL)
	}()

	origArgs := os.Args
	os.Args = []string{"app", "-a", ":8888", "-b", "http://flag.baseURL"}
	defer func() {
		os.Args = origArgs
	}()

	cfg := NewConfig()

	assert.Equal(t, DefaultServerAddress, cfg.ServerAddress)
	assert.Equal(t, "http://ex.ru", cfg.BaseURL)
}

func TestConfig_UsesFlagDefaultsWhenNothingSet(t *testing.T) {
	origServerAddress := os.Getenv("SERVER_ADDRESS")
	origBaseURL := os.Getenv("BASE_URL")

	os.Setenv("SERVER_ADDRESS", "")
	os.Setenv("BASE_URL", "")

	defer func() {
		os.Setenv("SERVER_ADDRESS", origServerAddress)
		os.Setenv("BASE_URL", origBaseURL)
	}()

	origArgs := os.Args
	os.Args = []string{"app"}
	defer func() {
		os.Args = origArgs
	}()

	cfg := NewConfig()

	assert.Equal(t, DefaultServerAddress, cfg.ServerAddress)
	assert.Equal(t, DefaultBaseURL, cfg.BaseURL)
}

func TestConfig_BaseURL_FormedFromServerAddress(t *testing.T) {
	origServerAddress := os.Getenv("SERVER_ADDRESS")
	origBaseURL := os.Getenv("BASE_URL")

	os.Setenv("SERVER_ADDRESS", ":8080")
	os.Setenv("BASE_URL", "")

	defer func() {
		os.Setenv("SERVER_ADDRESS", origServerAddress)
		os.Setenv("BASE_URL", origBaseURL)
	}()

	origArgs := os.Args
	os.Args = []string{"app", "-b", ""}
	defer func() {
		os.Args = origArgs
	}()

	cfg := NewConfig()

	assert.Equal(t, DefaultServerAddress, cfg.ServerAddress)
	assert.Equal(t, fmt.Sprintf("http://%s", DefaultServerAddress), cfg.BaseURL)
}