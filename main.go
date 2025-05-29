package main

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/cmd"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"go.uber.org/zap"
	"os"
)

var version string

var defaultMetricPort = "8080"

func main() {
	metricPort := os.Getenv("METRIC_PORT")
	if metricPort == "" {
		metricPort = defaultMetricPort
	}
	logger.InitDefaultLogger()
	defer func() { _ = zap.L().Sync() }()
	system.WithLogger(zap.L())
	zap.L().Debug("Start app", zap.String("version", version))
	go func() {
		defer system.HandlePanic()
		if err := metrics.Handler(metricPort); err != nil {
			zap.L().Fatal("error start metric", zap.Error(err))
		}
	}()
	cmd.Execute()
}
