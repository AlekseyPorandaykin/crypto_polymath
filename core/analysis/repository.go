package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

type Repository interface {
	Save(ctx context.Context, data Analytic) error
	Last(
		ctx context.Context,
		exchangeName, symbol string,
		unit domain.Unit,
		interval int,
		name string,
		indicatorDepth, depth int,
	) ([]Analytic, error)
}
