package price

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type StorageDTO struct {
	ID        uuid.UUID `db:"id"`
	Symbol    string    `db:"symbol"`
	Exchange  string    `db:"exchange"`
	Value     float64   `db:"value"`
	CreatedAt time.Time `db:"created_at"`
}

type Repository interface {
	Save(ctx context.Context, data ...StorageDTO) error
	Find(ctx context.Context, exchange, symbol string) (*StorageDTO, error)
	ListByExchange(ctx context.Context, exchange string) ([]StorageDTO, error)
	ListBySymbol(ctx context.Context, symbol string) ([]StorageDTO, error)
	Delete(ctx context.Context, exchange string, to time.Time) error
}
