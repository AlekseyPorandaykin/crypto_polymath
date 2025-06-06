package main

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/cmd"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
)

var version string

func main() {
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
