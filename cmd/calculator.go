package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/cmd/container"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/system"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/AlekseyPorandaykin/go-template/pkg/profiling"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os/signal"
	"runtime/debug"
	"syscall"
)

var calculatorCmd = &cobra.Command{
	Use:   "calculator",
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
		//Create applications
		calculationApp, errCreateCalculator := c.CreateCalculator()
		if errCreateCalculator != nil {
			logger.Error("error create calculator app", zap.Error(errCreateCalculator))
			return
		}
		//Create dispatchers
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
		loadedPriceDispatcher, errLoadedPriceDispatcher := c.CreateLoadedPricesDispatcher()
		if errLoadedPriceDispatcher != nil {
			logger.Error("error create loadedPriceDispatcher", zap.Error(errLoadedPriceDispatcher))
			return
		}
		candleIndicatorsDispatcher, errCandleIndicatorsDispatcher := c.CreateCandleIndicatorsDispatcher()
		if errCandleIndicatorsDispatcher != nil {
			logger.Error("error create errCandleIndicatorsDispatcher", zap.Error(errCandleIndicatorsDispatcher))
			return
		}

		//Run applications
		system.Go(func() {
			defer cancel()
			if err := calculationApp.Run(ctx); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run calculator app", zap.Error(err))
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
		system.Go(func() {
			loadedPriceDispatcher.Listen()
		})
		system.Go(func() {
			candleIndicatorsDispatcher.Listen()
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
	daemonCmd.AddCommand(calculatorCmd)
}
