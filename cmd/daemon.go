package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/cmd/container"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/logger"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os/signal"
	"runtime/debug"
	"syscall"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run server",
	Run: func(cmd *cobra.Command, args []string) {
		//Init global configs
		debug.SetMemoryLimit(400*2 ^ 20)
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
			fmt.Println("error init container", errLogger.Error())
			return
		}
		defer c.Close()
		loaderApp, errCreateLoader := c.CreateLoader()
		if errCreateLoader != nil {
			logger.Error("error create loader app", zap.Error(errCreateLoader))
			return
		}
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
		serverGrpc, errServerGrpc := c.CreateGRPCServer()
		if errServerGrpc != nil {
			logger.Error("error create grpc server", zap.Error(errServerGrpc))
			return
		}
		candleDispatcher, errCandleDispatcher := c.CreateCandleDispatcher()
		if errCandleDispatcher != nil {
			logger.Error("error create candle dispatcher", zap.Error(errCandleDispatcher))
			return
		}
		indicatorDispatcher, errIndicatorDispatcher := c.CreateIndicatorDispatcher()
		if errIndicatorDispatcher != nil {
			logger.Error("error create indicator dispatcher", zap.Error(errIndicatorDispatcher))
			return
		}
		createIndicatorDispatcher, errCreateIndicatorDispatcher := c.CreateCreatorIndicatorDispatcher()
		if errCreateIndicatorDispatcher != nil {
			logger.Error("error create creater indicator dispatcher", zap.Error(errCreateIndicatorDispatcher))
			return
		}
		analyticDispatcher, errAnalyticDispatcher := c.CreateAnalyticDispatcher()
		if errAnalyticDispatcher != nil {
			logger.Error("error create analytic dispatcher", zap.Error(errAnalyticDispatcher))
			return
		}
		loadedCandlestickDispatcher, errLoadedCandlestick := c.CreateLoadedCandlesticksForSymbolBodyDispatcher()
		if errLoadedCandlestick != nil {
			logger.Error("error create loadedCandlestickDispatcher", zap.Error(errLoadedCandlestick))
			return
		}

		//Run applications
		system.Go(func() {
			defer cancel()
			if err := loaderApp.Run(ctx); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run loader app", zap.Error(err))
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
			defer cancel()
			if err := serverHttp.Run(viper.GetString("http.host"), viper.GetString("http.port")); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run http server", zap.Error(err))
				return
			}
		})
		system.Go(func() {
			if err := serverGrpc.Run(ctx); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run grpc server", zap.Error(err))
				return
			}
		})
		system.Go(func() {
			candleDispatcher.Listen()
		})
		system.Go(func() {
			indicatorDispatcher.Listen()
		})
		system.Go(func() {
			createIndicatorDispatcher.Listen()
		})
		system.Go(func() {
			analyticDispatcher.Listen()
		})
		system.Go(func() {
			loadedCandlestickDispatcher.Listen()
		})

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
