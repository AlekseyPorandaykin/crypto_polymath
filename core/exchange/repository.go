package exchange

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type SymbolInfoStorageDTO struct {
	ID         uuid.UUID `db:"id"`
	Exchange   string    `db:"exchange"`
	Symbol     string    `db:"symbol"`
	BaseAsset  string    `db:"base_asset"`
	QuoteAsset string    `db:"quote_asset"`
	CreatedAt  time.Time `db:"created_at"`
}

type Repository interface {
	SaveSymbolInfo(ctx context.Context, data []SymbolInfoStorageDTO) error
	InfoBySymbol(ctx context.Context, exchange, symbol string) (*SymbolInfoStorageDTO, error)
	DeleteOldRows(ctx context.Context, exchangeName string, to time.Time) error
	ListByExchange(ctx context.Context, exchangeName string) ([]SymbolInfoStorageDTO, error)
}
