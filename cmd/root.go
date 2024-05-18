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

func Execute() {
	if err := rootCmd.Execute(); err != nil && !errors.Is(err, context.Canceled) {
		zap.L().Error("execute root cmd", zap.Error(err))
	}
}
