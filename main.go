package main

import (
	"strings"

	"github.com/AlekseyPorandaykin/crypto_polymath/cmd"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-kit/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version string

func main() {
	_ = godotenv.Load() // загружает .env в os environment
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	logger.InitDefaultLogger()
	defer func() { _ = zap.L().Sync() }()
	system.WithLogger(zap.L())
	zap.L().Debug("Start app", zap.String("version", version))
	go func() {
		defer system.HandlePanic()
		if err := metrics.Handler(viper.GetString("metric.port")); err != nil {
			zap.L().Fatal("error start metric", zap.Error(err))
		}
	}()
	cmd.Execute()
}
