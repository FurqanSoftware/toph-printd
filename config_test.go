package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.Token = "test-token"
		cfg.Toph.ContestID = "6502f105025832238e865526"

		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("missing token", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.ContestID = "6502f105025832238e865526"

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing token")
	})

	t.Run("missing contest ID", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.Token = "test-token"

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing contest ID")
	})

	t.Run("missing all", func(t *testing.T) {
		cfg := Config{}

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing Toph base URL")
		assert.Contains(t, err.Error(), "missing token")
		assert.Contains(t, err.Error(), "missing contest ID")
	})
}
