package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/laretz/internal/impl/plugin"
	"github.com/nikimonax/laretz/internal/shared/logging"
	"github.com/nikimonax/laretz/internal/shared/validator"
)

const AppName = "laretz"

const (
	DefaultServerHost = "127.0.0.1"
	DefaultServerPort = 8080
)

type ServerConfig struct {
	Port        uint16 `validate:"required"`
	Host        string `validate:"required"`
	OpenBrowser bool
}

func (cfg ServerConfig) Listen() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func (cfg ServerConfig) BrowserOpenURL() string {
	return "http://" + cfg.Listen() + "/"
}

type DataConfig struct {
	Dir string `validate:"required"`
}

func (cfg DataConfig) DatabaseFile() string {
	return filepath.Join(cfg.Dir, AppName+".db")
}

type LogConfig struct {
	Env   logging.LogEnv
	Level zapcore.Level
}

type Config struct {
	Server  ServerConfig
	Log     LogConfig
	Data    DataConfig
	Plugins map[plugin.PluginID]RawMessage
}

func DefaultAppDataDir() string {
	configDir, err := os.UserConfigDir()

	if err != nil {
		panic(err)
	}

	return filepath.Join(configDir, AppName)
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultAppDataDir(), "config.yaml")
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        DefaultServerPort,
			Host:        DefaultServerHost,
			OpenBrowser: false,
		},
		Data: DataConfig{
			Dir: DefaultAppDataDir(),
		},
	}
}

func ReadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, NewErrConfig("empty path")
	}

	configPath = filepath.Clean(configPath)

	_, err := os.Stat(configPath)

	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	cfg, err := DecodeConfig(content)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func WriteConfig(cfg *Config, configPath string) error {
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	content, err := EncodeConfig(cfg)

	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return err
	}

	return nil
}

func EnsureConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, NewErrConfig("empty path")
	}

	configPath = filepath.Clean(configPath)

	cfg, err := ReadConfig(configPath)

	if os.IsNotExist(err) {
		cfg = DefaultConfig()
		err = WriteConfig(cfg, configPath)
	}

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func ValidateConfig(cfg *Config) error {
	validate := validator.Get()
	err := validate.Struct(cfg)
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrConfig, err)
	}
	return err
}
