package container

import (
	"github.com/AlekseyPorandaykin/go-template/pkg/logger"
	"go.uber.org/zap"
)

type AppLogger *zap.Logger

type RepositoryLogger *zap.Logger

type ExchangeLogger *zap.Logger

type LoaderLogger *zap.Logger

type CalculatorLogger *zap.Logger

type HTTPServerLogger *zap.Logger
type HTTPClientLogger *zap.Logger

func asZapLogger[L ~*zap.Logger](l L) *zap.Logger {
	return (*zap.Logger)(l)
}

func (c *Container) initLoggers() error {
	if err := c.di.Provide(func() (AppLogger, error) {
		return logger.CreateForNamespace("")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (RepositoryLogger, error) {
		return logger.CreateForNamespace("repository")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (ExchangeLogger, error) {
		return logger.CreateForNamespace("exchange")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (LoaderLogger, error) {
		return logger.CreateForNamespace("loader")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (CalculatorLogger, error) {
		return logger.CreateForNamespace("calculator")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (HTTPServerLogger, error) {
		return logger.CreateForNamespace("http_server")
	}); err != nil {
		return err
	}
	if err := c.di.Provide(func() (HTTPClientLogger, error) {
		return logger.CreateForNamespace("http_client")
	}); err != nil {
		return err
	}
	return nil
}
