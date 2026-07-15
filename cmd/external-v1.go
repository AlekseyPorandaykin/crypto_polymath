package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/AlekseyPorandaykin/crypto_polymath/cmd/container"
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"github.com/AlekseyPorandaykin/go-template/pkg/profiling"
	"github.com/AlekseyPorandaykin/go-template/pkg/system"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var externalV1Cmd = &cobra.Command{
	Use:   "external-v1",
	Short: "Run external v1 server",
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
			fmt.Println("error init container", err.Error())
			return
		}
		defer c.Close()
		serverHttp, errCreateServer := c.CreateApiServer()
		if errCreateServer != nil {
			logger.Error("error create http server", zap.Error(errCreateServer))
			return
		}

		//Run applications
		system.Go(func() {
			defer cancel()
			if err := serverHttp.Run(viper.GetString("http.host"), viper.GetString("http.port")); !errors.Is(err, context.Canceled) && err != nil {
				logger.Info("run http server", zap.Error(err))
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
	apiCmd.AddCommand(externalV1Cmd)
}
