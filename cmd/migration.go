package cmd

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/config"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/database"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Migration database",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		cfg := config.Create()

		conn, err := database.CreateConnection(cfg.DBConnection)
		if err != nil {
			zap.L().Error("error create connection to db", zap.Error(err))
			return
		}
		defer func() { _ = conn.Close() }()
		queries, err := specification()
		if err != nil {
			return
		}
		for name, query := range queries {
			_, errExec := conn.ExecContext(ctx, query)
			if errExec != nil {
				fmt.Println("error execute migration", name, errExec.Error())
				continue
			}
			fmt.Println("execute migration", name)
		}
		cancel()
	},
}

func specification() (map[string]string, error) {
	dirName := "./migrations/sqlite"
	dirs, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	queries := make(map[string]string, len(dirs))
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		path := filepath.Join(dirName, dir.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			continue
		}
		queries[dir.Name()] = string(data)
	}

	return queries, nil
}

func init() {
	rootCmd.AddCommand(migrationCmd)
}
