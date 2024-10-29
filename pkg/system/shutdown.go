package system

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger = zap.L()
}

func WithLogger(l *zap.Logger) {
	if l == nil {
		return
	}
	logger = l
}

func HandlePanic() {
	if err := recover(); err != nil {
		logger.Error("handle panic", zap.Any("recover", err))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func Go(run func()) {
	go func() {
		defer HandlePanic()
		run()
	}()
}

func MustInit[T any](res T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("error init: %s", err.Error()))
	}
	return res
}
