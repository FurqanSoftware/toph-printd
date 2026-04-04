package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.Token = "keyboardcat"
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
		cfg.Toph.Token = "keyboardcat"

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

	t.Run("invalid contest ID", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.Token = "keyboardcat"
		cfg.Toph.ContestID = "not-a-valid-id"

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid contest ID")
	})

	t.Run("contest ID too short", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.Token = "keyboardcat"
		cfg.Toph.ContestID = "6502f105"

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid contest ID")
	})

	t.Run("contest ID uppercase rejected", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.Token = "keyboardcat"
		cfg.Toph.ContestID = "6502F105025832238E865526"

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid contest ID")
	})

	t.Run("HTTP base URL rejected", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.BaseURL = "http://toph.co"
		cfg.Toph.Token = "keyboardcat"
		cfg.Toph.ContestID = "6502f105025832238e865526"

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Toph base URL must use HTTPS")
	})

	t.Run("HTTP allowed for .local", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.BaseURL = "http://toph.local"
		cfg.Toph.Token = "keyboardcat"
		cfg.Toph.ContestID = "6502f105025832238e865526"

		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("HTTPS base URL accepted", func(t *testing.T) {
		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.BaseURL = "https://toph.co"
		cfg.Toph.Token = "keyboardcat"
		cfg.Toph.ContestID = "6502f105025832238e865526"

		err := validateConfig(cfg)
		assert.NoError(t, err)
	})
}
