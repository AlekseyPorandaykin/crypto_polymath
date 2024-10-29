package logger

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type Config struct {
	Level            string   `mapstructure:"level"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	Stacktrace       bool     `mapstructure:"stacktrace"`
}

var DefaultConf = Config{Level: "DEBUG", OutputPaths: []string{"stdout"}}

func Create(conf Config, opts ...zap.Option) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(conf.Level)
	if err != nil {
		return nil, err
	}
	l, err := zap.Config{
		Level:             level,
		Development:       false,
		Encoding:          "json",
		DisableStacktrace: !conf.Stacktrace,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "ts",
			LevelKey:    "level",
			MessageKey:  "message",
			LineEnding:  zapcore.DefaultLineEnding,
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime:  zapcore.TimeEncoderOfLayout(time.DateTime),
		},
		OutputPaths:      conf.OutputPaths,
		ErrorOutputPaths: conf.ErrorOutputPaths,
	}.Build(opts...)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func CreateForNamespace(namespace string) (*zap.Logger, error) {
	l, err := Create(createNamespaceConfig(namespace))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("create logger namesapce: %s", namespace))
	}
	return l.Named(namespace), nil
}

func createNamespaceConfig(namespace string) Config {
	conf := DefaultConf
	keyLevel := "logger.level"
	keyOutputPaths := "logger.output_paths"
	keyErrorOutputPaths := "logger.error_output_paths"
	keyStacktrace := "logger.stacktrace"
	if namespace != "" {
		keyLevel = fmt.Sprintf("%s.%s", namespace, keyLevel)
		keyOutputPaths = fmt.Sprintf("%s.%s", namespace, keyOutputPaths)
		keyErrorOutputPaths = fmt.Sprintf("%s.%s", namespace, keyErrorOutputPaths)
		keyStacktrace = fmt.Sprintf("%s.%s", namespace, keyStacktrace)
	}
	level := viper.GetString(keyLevel)
	outputPaths := viper.GetStringSlice(keyOutputPaths)
	errorOutputPaths := viper.GetStringSlice(keyErrorOutputPaths)
	conf.Stacktrace = viper.GetBool(keyStacktrace)
	if level != "" {
		conf.Level = level
	}
	if len(outputPaths) > 0 {
		conf.OutputPaths = outputPaths
	}
	if len(errorOutputPaths) > 0 {
		conf.ErrorOutputPaths = errorOutputPaths
	}

	return conf
}

func CreateGlobal(conf Config, opts ...zap.Option) {
	zap.ReplaceGlobals(zap.Must(Create(conf, opts...)))
}

func InitDefaultLogger() {
	CreateGlobal(createNamespaceConfig(""))
}
