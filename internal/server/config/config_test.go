package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/server/config"
	"github.com/nikimonax/laretz/internal/shared/logging"
)

func TestDecodeConfigRejectsUnknownField(t *testing.T) {
	_, err := config.DecodeConfig([]byte("server:\n  port: 8080\n  listen: 127.0.0.1\n"))

	require.Error(t, err, "unknown keys must fail loudly, not be silently ignored")
	require.ErrorIs(t, err, config.ErrConfig)
	assert.Contains(t, err.Error(), "listen")
}

func TestDecodeConfigFull(t *testing.T) {
	content := []byte(`server:
  port: 8080
  host: 127.0.0.1
  openbrowser: false

log:
  env: dev
  level: info

data:
  dir: /tmp/laretz

plugins:
  coingecko:
    api_key: test-key
`)

	cfg, err := config.DecodeConfig(content)

	require.NoError(t, err)
	assert.Equal(t, 8080, int(cfg.Server.Port))
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, logging.EnvDev, cfg.Log.Env)
	assert.Equal(t, "/tmp/laretz", cfg.Data.Dir)
	assert.Contains(t, cfg.Plugins, "coingecko")
}

func TestConfigRoundTrip(t *testing.T) {
	cfg, err := config.DefaultConfig()

	require.NoError(t, err)

	encoded, err := config.EncodeConfig(cfg)

	require.NoError(t, err)
	assert.Contains(t, string(encoded), "env: dev", "env must serialize as a string, not a number")
	assert.Contains(t, string(encoded), "level: info")

	cfg, err = config.DecodeConfig(encoded)

	require.NoError(t, err, "config written by first run must be readable by second run")
	assert.Equal(t, logging.EnvDev, cfg.Log.Env)
	assert.Equal(t, config.DefaultServerHost, cfg.Server.Host)
	assert.Equal(t, config.DefaultServerPort, int(cfg.Server.Port))
}

func TestValidateConfig(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		cfg, err := config.DefaultConfig()

		require.NoError(t, err)

		assert.NoError(t, config.ValidateConfig(cfg))
	})

	t.Run("empty host is invalid", func(t *testing.T) {
		cfg, err := config.DefaultConfig()

		require.NoError(t, err)

		cfg.Server.Host = ""

		err = config.ValidateConfig(cfg)

		require.ErrorIs(t, err, config.ErrConfig)
	})

	t.Run("empty data dir is invalid", func(t *testing.T) {
		cfg, err := config.DefaultConfig()

		require.NoError(t, err)

		cfg.Data.Dir = ""

		err = config.ValidateConfig(cfg)

		require.ErrorIs(t, err, config.ErrConfig)
	})
}
