package metrics

import (
	"errors"
	"net"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(host, port string) error {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/readyz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(net.JoinHostPort(host, port), nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func EchoHandler(e *echo.Echo, subsystem, host, port string) func() error {
	e.Use(echoprometheus.NewMiddleware(subsystem))
	return func() error {
		metrics := echo.New()
		metrics.GET("/healthz", func(c echo.Context) error {
			return c.JSON(http.StatusOK, nil)
		})
		metrics.GET("/readyz", func(c echo.Context) error {
			return c.JSON(http.StatusOK, nil)
		})
		if err := metrics.Start(net.JoinHostPort(host, port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
