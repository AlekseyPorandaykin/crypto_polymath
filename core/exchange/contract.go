package exchange

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"time"
)

type SymbolCategory string

const (
	SymbolCategorySpot   SymbolCategory = "spot"
	SymbolCategoryFuture SymbolCategory = "future"
	SymbolCategoryOther  SymbolCategory = "other"
)

type SymbolInfoDTO struct {
	Symbol          string
	Exchange        string
	BaseAsset       string
	QuoteAsset      string
	Category        SymbolCategory
	FundingRate     float32
	NextFundingTime time.Time
}

type ExternalLoader interface {
	SymbolInfo(ctx context.Context) ([]SymbolInfoDTO, error)
}

type Exchange interface {
	AddLoader(exchangeName string, loader ExternalLoader)
	LoadSymbolInfo(ctx context.Context, exchangeName string) ([]domain.SymbolInfo, error)
	SymbolInfo(ctx context.Context, exchangeName, symbol string) (*domain.SymbolInfo, error)
	SymbolInfoByCategory(ctx context.Context, exchange, category string) ([]domain.SymbolInfo, error)
}
