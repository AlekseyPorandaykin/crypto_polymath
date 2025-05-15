package logger

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var setLoggers = map[string]*zap.Logger{}

type Config struct {
	Namespace        string   `mapstructure:"namespace"`
	Level            string   `mapstructure:"level"`
	AlertLevel       string   `mapstructure:"alert_level"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	Stacktrace       bool     `mapstructure:"stacktrace"`
}

var DefaultConf = Config{Namespace: "App", Level: "DEBUG", OutputPaths: []string{"stdout"}}

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
	alertLevel := zap.ErrorLevel
	if conf.AlertLevel != "" {
		_ = alertLevel.Set(conf.AlertLevel)
	}
	sentryLog, err := modifyToSentryLogger(l, alertLevel, conf.Namespace)
	if err != nil {
		return nil, err
	}
	l = sentryLog

	//TODO: выбрать другой канал отправки
	//l = l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
	//	alertLevel := zap.ErrorLevel
	//	if conf.AlertLevel != "" {
	//		_ = alertLevel.Set(conf.AlertLevel)
	//	}
	//	return zapcore.NewTee(core, NewMailZapCore(
	//		alertLevel,
	//		viper.GetString("smtp_username"),
	//		viper.GetString("smtp_password"),
	//		viper.GetString("smtp_host"),
	//		viper.GetUint("smtp_port"),
	//		viper.GetString("smtp_from"),
	//		viper.GetString("error_mail_to")),
	//	)
	//}))
	if hostname, err := os.Hostname(); err == nil {
		l = l.With(zap.String("hostname", hostname))
	}
	return l, nil
}

func CreateForNamespace(namespace string) (*zap.Logger, error) {
	if logger, has := setLoggers[namespace]; has {
		return logger, nil
	}
	l, err := Create(createNamespaceConfig(namespace))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("create logger namesapce: %s", namespace))
	}
	setLoggers[namespace] = l
	return l.Named(namespace), nil
}

func createNamespaceConfig(namespace string) Config {
	conf := DefaultConf
	conf.Namespace = namespace
	keyLevel := "logger.level"
	keyAlertLevel := "logger.alert_level"
	keyOutputPaths := "logger.output_paths"
	keyErrorOutputPaths := "logger.error_output_paths"
	keyStacktrace := "logger.stacktrace"
	if namespace != "" {
		keyLevel = fmt.Sprintf("%s.%s", namespace, keyLevel)
		keyAlertLevel = fmt.Sprintf("%s.%s", namespace, keyAlertLevel)
		keyOutputPaths = fmt.Sprintf("%s.%s", namespace, keyOutputPaths)
		keyErrorOutputPaths = fmt.Sprintf("%s.%s", namespace, keyErrorOutputPaths)
		keyStacktrace = fmt.Sprintf("%s.%s", namespace, keyStacktrace)
	}
	level := viper.GetString(keyLevel)
	alertLevel := viper.GetString(keyAlertLevel)
	outputPaths := viper.GetStringSlice(keyOutputPaths)
	errorOutputPaths := viper.GetStringSlice(keyErrorOutputPaths)
	conf.Stacktrace = viper.GetBool(keyStacktrace)
	if level != "" {
		conf.Level = level
	}
	if alertLevel != "" {
		conf.AlertLevel = alertLevel
	}
	if len(outputPaths) > 0 {
		conf.OutputPaths = outputPaths
	}
	if len(errorOutputPaths) > 0 {
		conf.ErrorOutputPaths = errorOutputPaths
	}

	return conf
}

func CreateGlobal(conf Config) {
	l, has := setLoggers[conf.Namespace]
	if !has {
		l = zap.Must(Create(conf))
		setLoggers[conf.Namespace] = l
	}

	zap.ReplaceGlobals(l)
}

func InitDefaultLogger() {
	CreateGlobal(createNamespaceConfig(""))
}

func AttachCoreToLogger(core zapcore.Core, l *zap.Logger) *zap.Logger {
	return l.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, core)
	}))
}
