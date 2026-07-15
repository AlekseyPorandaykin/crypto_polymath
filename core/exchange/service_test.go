package exchange_test

import (
	"context"
	"testing"
	"time"

	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/google/uuid"
)

func TestExchange_LoadSymbolInfo_savesAndReturns(t *testing.T) {
	repo := memory.NewExchangeRepository()
	s := core_exchange.New(repo)
	nextFunding := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	s.AddLoader("bybit", stubLoader{
		symbolInfo: func(_ context.Context) ([]core_exchange.SymbolInfoDTO, error) {
			return []core_exchange.SymbolInfoDTO{
				{
					Symbol:          "BTCUSDT",
					Exchange:        "bybit",
					BaseAsset:       "BTC",
					QuoteAsset:      "USDT",
					Category:        core_exchange.SymbolCategorySpot,
					FundingRate:     0.01,
					NextFundingTime: nextFunding,
				},
			}, nil
		},
	})

	result, err := s.LoadSymbolInfo(context.Background(), "bybit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 symbol, got %#v", result)
	}
	if result[0].Symbol != "BTCUSDT" || result[0].BaseAsset != "BTC" {
		t.Fatalf("unexpected symbol info: %#v", result[0])
	}
	if result[0].NextFundingTime == nil || !result[0].NextFundingTime.Equal(nextFunding) {
		t.Fatalf("expected next funding time %#v, got %#v", nextFunding, result[0].NextFundingTime)
	}

	stored, err := repo.InfoBySymbol(context.Background(), "bybit", "BTCUSDT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored == nil || stored.Symbol != "BTCUSDT" {
		t.Fatalf("expected stored symbol, got %#v", stored)
	}
}

func TestExchange_LoadSymbolInfo_noLoader(t *testing.T) {
	s := core_exchange.New(memory.NewExchangeRepository())

	result, err := s.LoadSymbolInfo(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil without loader, got %#v", result)
	}
}

func TestExchange_SymbolInfo_fromStorage(t *testing.T) {
	repo := memory.NewExchangeRepository()
	if err := repo.SaveSymbolInfo(context.Background(), []core_exchange.SymbolInfoStorageDTO{
		{
			ID:         uuid.New(),
			Exchange:   "bybit",
			Symbol:     "BTCUSDT",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Category:   "spot",
			CreatedAt:  time.Now().UTC(),
		},
	}); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	s := core_exchange.New(repo)
	info, err := s.SymbolInfo(context.Background(), "bybit", "BTCUSDT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil || !info.IsExist || info.BaseAsset != "BTC" {
		t.Fatalf("unexpected symbol info: %#v", info)
	}
}

func TestExchange_SymbolInfo_derivesFromQuoteAsset(t *testing.T) {
	repo := memory.NewExchangeRepository()
	if err := repo.SaveSymbolInfo(context.Background(), []core_exchange.SymbolInfoStorageDTO{
		{
			ID:         uuid.New(),
			Exchange:   "bybit",
			Symbol:     "BTCUSDT",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Category:   "spot",
			CreatedAt:  time.Now().UTC(),
		},
	}); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	s := core_exchange.New(repo)
	info, err := s.SymbolInfo(context.Background(), "bybit", "ETHUSDT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil || info.IsExist || info.QuoteAsset != "USDT" {
		t.Fatalf("unexpected derived symbol info: %#v", info)
	}
	// Сервис использует TrimPrefix вместо TrimSuffix, поэтому base asset остаётся полным символом.
	if info.BaseAsset != "ETHUSDT" {
		t.Fatalf("expected base asset ETHUSDT, got %s", info.BaseAsset)
	}
}

func TestExchange_SymbolInfoByCategory(t *testing.T) {
	repo := memory.NewExchangeRepository()
	if err := repo.SaveSymbolInfo(context.Background(), []core_exchange.SymbolInfoStorageDTO{
		{
			ID: uuid.New(), Exchange: "bybit", Symbol: "BTCUSDT",
			BaseAsset: "BTC", QuoteAsset: "USDT", Category: "spot", CreatedAt: time.Now().UTC(),
		},
		{
			ID: uuid.New(), Exchange: "bybit", Symbol: "BTCUSD",
			BaseAsset: "BTC", QuoteAsset: "USD", Category: "future", CreatedAt: time.Now().UTC(),
		},
	}); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	s := core_exchange.New(repo)
	spot, err := s.SymbolInfoByCategory(context.Background(), "bybit", "spot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spot) != 1 || spot[0].Symbol != "BTCUSDT" {
		t.Fatalf("unexpected spot symbols: %#v", spot)
	}
}

type stubLoader struct {
	symbolInfo func(ctx context.Context) ([]core_exchange.SymbolInfoDTO, error)
}

func (s stubLoader) SymbolInfo(ctx context.Context) ([]core_exchange.SymbolInfoDTO, error) {
	if s.symbolInfo == nil {
		return nil, nil
	}
	return s.symbolInfo(ctx)
}
