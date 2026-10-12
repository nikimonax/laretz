package logging

import (
	"errors"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEnv int

const (
	EnvDev LogEnv = iota
	EnvProd
)

var (
	ErrUnknownEnv = errors.New("unknown env")
)

func (env LogEnv) String() string {
	switch env {
	case EnvDev:
		return "dev"
	case EnvProd:
		return "prod"
	default:
		panic(ErrUnknownEnv)
	}
}

func (env LogEnv) MarshalText() ([]byte, error) {
	return []byte(env.String()), nil
}

func EnvFromString(env string) (LogEnv, error) {
	switch env {
	case "dev":
		return EnvDev, nil
	case "prod":
		return EnvProd, nil
	default:
		return LogEnv(0), ErrUnknownEnv
	}
}

func NewZapLogger(env LogEnv, level zapcore.Level) *zap.Logger {
	var config zap.Config

	switch env {
	case EnvDev:
		config = zap.NewDevelopmentConfig()
	case EnvProd:
		config = zap.NewProductionConfig()
	default:
		panic(ErrUnknownEnv)
	}

	config.DisableCaller = true
	config.Level.SetLevel(level)

	logger, err := config.Build()

	if err != nil {
		log.Fatalf("failed create zap logger: %s", err)
	}

	return logger
}
