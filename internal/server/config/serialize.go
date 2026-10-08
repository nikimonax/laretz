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
	yaml.RegisterCustomMarshaler(marshalZapcoreLevel)

	yaml.RegisterCustomUnmarshaler(unmarshalZapEnv)
	yaml.RegisterCustomMarshaler(marshalZapEnv)
}

var (
	configMarshal   = yaml.Marshal
	configUnmarshal = yaml.Unmarshal
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

func marshalZapcoreLevel(level *zapcore.Level) ([]byte, error) {
	return []byte(level.String()), nil
}

func unmarshalZapEnv(env *logging.LogEnv, data []byte) (err error) {
	*env, err = logging.EnvFromString(strings.TrimSpace(string(data)))
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrConfig, err)
	}
	return err
}

func marshalZapEnv(env *logging.LogEnv) ([]byte, error) {
	return []byte(env.String()), nil
}

func EncodeConfig(cfg *Config) ([]byte, error) {
	return configMarshal(cfg)
}

func DecodeConfig(content []byte) (*Config, error) {
	cfg := new(Config)

	if err := configUnmarshal(content, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
