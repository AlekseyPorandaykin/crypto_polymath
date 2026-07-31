package container

import (
	"net/http"
	"testing"

	"github.com/AlekseyPorandaykin/crypto-exchanges/client"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/adapters"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func TestBybitExchangeClientSingleton(t *testing.T) {
	viper.Set("bybit.host", "https://api.bybit.com/")

	c := &Container{di: dig.New()}
	if err := c.di.Provide(func() (ExchangeLogger, error) {
		return zap.NewNop(), nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := c.di.Provide(func(_ ExchangeLogger) http.RoundTripper {
		return http.DefaultTransport
	}); err != nil {
		t.Fatal(err)
	}
	if err := c.initExchanges(); err != nil {
		t.Fatal(err)
	}

	err := c.di.Invoke(func(first, second client.ExchangeClient) {
		if first != second {
			t.Fatal("client.ExchangeClient must be a singleton in DI")
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	err = c.di.Invoke(func(first, second *adapters.Exchange) {
		if first != second {
			t.Fatal("*adapters.Exchange must be a singleton in DI")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
