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
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var calculatorCmd = &cobra.Command{
	Use:   "calculator",
	Short: "Run server",
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

		//Run applications
		system.Go(func() {
			defer cancel()
			if err := calculationApp.Run(ctx); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run calculator app", zap.Error(err))
				return
			}
		})

		//system.Go(func() {
		//	if err := profiling.StartPprofServer("0.0.0.0:6060"); err != nil {
		//		logger.Error("error start pprof server", zap.Error(err))
		//		return
		//	}
		//})

		<-ctx.Done()
	},
}

func init() {
	daemonCmd.AddCommand(calculatorCmd)
}
