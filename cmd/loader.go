package cmd

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/cmd/container"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/AlekseyPorandaykin/go-template/pkg/profiling"
	"github.com/AlekseyPorandaykin/go-template/pkg/system"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os/signal"
	"runtime/debug"
	"syscall"
)

var loaderCmd = &cobra.Command{
	Use:   "loader",
	Short: "Run loader daemon",
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
			if err := loader.Run(ctx); err != nil {
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
			if err := profiling.StartPprofServer("0.0.0.0:6060"); err != nil {
				logger.Error("error start pprof server", zap.Error(err))
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

		<-ctx.Done()
	},
}

func init() {
	daemonCmd.AddCommand(loaderCmd)
}
