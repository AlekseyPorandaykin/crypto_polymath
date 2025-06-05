package cmd

import (
	"context"
	"errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var homeDir = "./"

var rootCmd = &cobra.Command{
	Use: "crypto_polymath",
}

var daemonCmd = &cobra.Command{
	Use: "daemon",
}

var apiCmd = &cobra.Command{
	Use: "api",
}

var scriptCmd = &cobra.Command{
	Use: "script",
}

func init() {
	rootCmd.AddCommand(daemonCmd, apiCmd, scriptCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil && !errors.Is(err, context.Canceled) {
		zap.L().Error("execute root cmd", zap.Error(err))
	}
}
