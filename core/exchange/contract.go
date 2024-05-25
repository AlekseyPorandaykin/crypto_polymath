package exchange

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

type SymbolInfoDTO struct {
	Symbol     string
	Exchange   string
	BaseAsset  string
	QuoteAsset string
}

type ExternalLoader interface {
	SymbolInfo(ctx context.Context) ([]SymbolInfoDTO, error)
}

type Exchange interface {
	AddLoader(exchangeName string, loader ExternalLoader)
	LoadSymbolInfo(ctx context.Context, exchangeName string) ([]domain.SymbolInfo, error)
	SymbolInfo(ctx context.Context, exchangeName, symbol string) (*domain.SymbolInfo, error)
}
