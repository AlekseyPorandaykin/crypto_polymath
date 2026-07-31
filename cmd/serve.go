package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/AlekseyPorandaykin/crypto_polymath/cmd/container"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-kit/pkg/logger"
	"github.com/AlekseyPorandaykin/go-kit/pkg/profiling"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run all applications",
	Run: func(cmd *cobra.Command, args []string) {
		//Init global configs
		debug.SetMemoryLimit(400 * 1024 * 1024)
		debug.SetGCPercent(30)

		defer system.HandlePanic()
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		logger, errLogger := logger.CreateForNamespace("daemon")
		if errLogger != nil {
			fmt.Println("error create logger", errLogger.Error())
			return
		}
		defer func() { _ = logger.Sync() }()
		c := container.NewContainer()
		if err := c.Init(); err != nil {
			fmt.Println("error init container", err.Error())
			return
		}
		defer c.Close()
		//Create applications
		calculationApp, errCreateCalculator := c.CreateCalculator()
		if errCreateCalculator != nil {
			logger.Error("error create calculator app", zap.Error(errCreateCalculator))
			return
		}
		serverHttp, errCreateServer := c.CreateApiServer()
		if errCreateServer != nil {
			logger.Error("error create http server", zap.Error(errCreateServer))
			return
		}
		loader, errLoader := c.CreateLoader()
		if errLoader != nil {
			logger.Error("error create loader", zap.Error(errLoader))
			return
		}
		grpcServer, errGrpcServer := c.CreateGRPCServer()
		if errGrpcServer != nil {
			logger.Error("error create grpc server", zap.Error(errGrpcServer))
			return
		}
		loadedCandlesticksDispatcher, errLoadedCandlesticksDispatcher := c.CreateLoadedCandlesticksForSymbolBodyDispatcher()
		if errLoadedCandlesticksDispatcher != nil {
			logger.Error("error create loaded candlesticks dispatcher", zap.Error(errLoadedCandlesticksDispatcher))
			return
		}
		candleDispatcher, errCandleDispatcher := c.CreateCandleDispatcher()
		if errCandleDispatcher != nil {
			logger.Error("error create candle dispatcher", zap.Error(errCandleDispatcher))
			return
		}
		loadedPricesDispatcher, errLoadedPricesDispatcher := c.CreateLoadedPricesDispatcher()
		if errLoadedPricesDispatcher != nil {
			logger.Error("error create loaded prices dispatcher", zap.Error(errLoadedPricesDispatcher))
			return
		}

		//Run applications
		system.Go(func() {
			defer cancel()
			if err := loader.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("error start loader", zap.Error(err))
				return
			}
		})
		system.Go(func() {
			defer cancel()
			if err := grpcServer.Run(ctx); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run grpcServer app", zap.Error(err))
				return
			}
		})
		system.Go(func() {
			loadedCandlesticksDispatcher.Listen()
		})
		system.Go(func() {
			candleDispatcher.Listen()
		})
		system.Go(func() {
			loadedPricesDispatcher.Listen()
		})
		system.Go(func() {
			defer cancel()
			if err := serverHttp.Run(viper.GetString("http.host"), viper.GetString("http.port")); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run http server", zap.Error(err))
				return
			}
		})
		system.Go(func() {
			defer cancel()
			if err := calculationApp.Run(ctx); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run calculator app", zap.Error(err))
				return
			}
		})

		system.Go(func() {
			if err := profiling.StartPprofServer("0.0.0.0:6060"); err != nil {
				logger.Error("error start pprof server", zap.Error(err))
				return
			}
		})

		<-ctx.Done()
	},
}

func init() {
	daemonCmd.AddCommand(serveCmd)
}
