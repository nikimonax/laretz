package config

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/laretz/internal/shared/logging"
)

func init() {
	yaml.RegisterCustomUnmarshaler(unmarshalZapcoreLevel)
	yaml.RegisterCustomUnmarshaler(unmarshalZapEnv)
}

var (
	configMarshal   = yaml.Marshal
	configUnmarshal = yaml.UnmarshalWithOptions
)

type RawMessage struct {
	unmarshal func(any) error
}

func (msg *RawMessage) UnmarshalYAML(unmarshal func(any) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *RawMessage) Unmarshal(v any) error {
	return msg.unmarshal(v)
}

func unmarshalZapcoreLevel(level *zapcore.Level, data []byte) (err error) {
	*level, err = zapcore.ParseLevel(strings.TrimSpace(string(data)))
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrConfig, err)
	}
	return err
}

func unmarshalZapEnv(env *logging.LogEnv, data []byte) (err error) {
	*env, err = logging.EnvFromString(strings.TrimSpace(string(data)))
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrConfig, err)
	}
	return err
}

func EncodeConfig(cfg *Config) ([]byte, error) {
	return configMarshal(cfg)
}

func DecodeConfig(content []byte) (*Config, error) {
	cfg := new(Config)

	if err := configUnmarshal(content, cfg, yaml.DisallowUnknownField()); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfig, err)
	}

	return cfg, nil
}
