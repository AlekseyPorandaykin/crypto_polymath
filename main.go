package main

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/cmd"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/logger"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"go.uber.org/zap"
)

var version string

func main() {
	logger.InitDefaultLogger()
	defer func() { _ = zap.L().Sync() }()
	system.WithLogger(zap.L())
	zap.L().Debug("Start app", zap.String("version", version))
	go func() {
		defer system.HandlePanic()
		if err := metrics.Handler("localhost", "8080"); err != nil {
			zap.L().Fatal("error start metric", zap.Error(err))
		}
	}()
	cmd.Execute()
}
