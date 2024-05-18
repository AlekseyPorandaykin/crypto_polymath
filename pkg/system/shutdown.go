package system

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func HandlePanic() {
	if err := recover(); err != nil {
		zap.L().Error("handle panic", zap.Any("recover", err))
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
