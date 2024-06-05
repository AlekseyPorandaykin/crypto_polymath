package sqlite

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type AnalyticDTO struct {
	ID             uuid.UUID `db:"id"`
	Name           string    `db:"name"`
	Exchange       string    `db:"exchange"`
	Symbol         string    `db:"symbol"`
	Unit           string    `db:"unit"`
	Interval       int       `db:"interval"`
	Datetime       time.Time `db:"datetime"`
	Depth          int       `db:"depth"`
	ByIndicator    string    `db:"by_indicator"`
	IndicatorDepth int       `db:"indicator_depth"`
	Value          float64   `db:"value"`
	CreatedAt      time.Time `db:"created_at"`
}

type AnalyticRepository struct {
	db *sqlx.DB
}

func NewAnalyticRepository(db *sqlx.DB) *AnalyticRepository {
	return &AnalyticRepository{db: db}
}

func (repo *AnalyticRepository) Save(ctx context.Context, data analysis.Analytic) error {
	defer metrics.CacheQueryHelper("crypto_polymath", "analysis_save")()
	var query = `
INSERT INTO analytics(
                      id, 
                      name, 
                      exchange, 
                      symbol, 
                      unit, 
                      interval, 
                      datetime, 
                      depth, 
                      by_indicator, 
                      indicator_depth, 
                      value,
                      created_at
)
VALUES (
           :id,
           :name,
           :exchange,
           :symbol,
           :unit,
           :interval,
           :datetime,
           :depth,
           :by_indicator,
           :indicator_depth,
           :value,
           :created_at
       )
`
	_, err := repo.db.NamedExecContext(ctx, query, AnalyticDTO{
		ID:             data.ID,
		Name:           data.Name,
		Exchange:       data.Exchange,
		Symbol:         data.Symbol,
		Unit:           string(data.Unit),
		Interval:       data.Interval,
		Datetime:       data.Datetime.In(time.UTC),
		Depth:          data.Depth,
		ByIndicator:    data.ByIndicator,
		IndicatorDepth: data.IndicatorDepth,
		Value:          data.Value,
		CreatedAt:      time.Now().In(time.UTC),
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *AnalyticRepository) Last(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int,
) ([]analysis.Analytic, error) {
	defer metrics.CacheQueryHelper("crypto_polymath", "analysis_last")()
	var query = `
SELECT id,
       name,
       exchange,
       symbol,
       unit,
       interval,
       datetime,
       depth,
       by_indicator,
       indicator_depth,
       value,
       created_at
FROM analytics
WHERE exchange = ?
  AND symbol = ?
  AND unit = ?
  AND interval = ?
  AND name = ?
  AND indicator_depth = ?
  AND depth = ?
ORDER BY datetime DESC
LIMIT 100
`
	data := make([]AnalyticDTO, 0, 100)
	err := repo.db.SelectContext(ctx, &data, query, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	analytics := make([]analysis.Analytic, 0, 100)
	for _, item := range data {
		analytics = append(analytics, analysis.Analytic{
			ID:             item.ID,
			Name:           item.Name,
			Exchange:       item.Exchange,
			Symbol:         item.Symbol,
			Unit:           domain.Unit(item.Unit),
			Interval:       item.Interval,
			Datetime:       item.Datetime,
			Depth:          item.Depth,
			ByIndicator:    item.ByIndicator,
			IndicatorDepth: item.IndicatorDepth,
			Value:          item.Value,
		})
	}
	return analytics, nil
}
