package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/view"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"math"
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
	if math.IsNaN(data.Value) {
		return nil
	}
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_save")()
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

func (repo *AnalyticRepository) Find(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	datetime time.Time,
	interval, indicatorDepth, depth int,
) (*analysis.Analytic, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_find")()
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
  AND datetime = ?
  AND name = ?
  AND indicator_depth = ?
  AND depth = ?
`
	storageData := AnalyticDTO{}
	err := repo.db.GetContext(
		ctx,
		&storageData,
		query,
		exchangeName,
		symbol,
		string(unit),
		interval,
		datetime,
		name,
		indicatorDepth,
		depth,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &analysis.Analytic{
		ID:             storageData.ID,
		Name:           storageData.Name,
		Exchange:       storageData.Exchange,
		Symbol:         storageData.Symbol,
		Unit:           domain.Unit(storageData.Unit),
		Interval:       storageData.Interval,
		Datetime:       storageData.Datetime,
		Depth:          storageData.Depth,
		ByIndicator:    storageData.ByIndicator,
		IndicatorDepth: storageData.IndicatorDepth,
		Value:          storageData.Value,
	}, nil
}

func (repo *AnalyticRepository) Last(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int,
) ([]analysis.Analytic, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_last")()
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

func (repo *AnalyticRepository) UniqGroups(ctx context.Context) ([]analysis.UniqGroup, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_uniq_groups")()
	var query = `
SELECT DISTINCT name,
       exchange,
       symbol,
       unit,
       interval,
       depth,
       by_indicator,
       indicator_depth
FROM analytics
`
	var data []analysis.UniqGroup
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			name, exchange, symbol, unit, byIndicator string
			interval, depth, indicatorDepth           int
		)
		if err := rows.Scan(&name, &exchange, &symbol, &unit, &interval, &depth, &byIndicator, &indicatorDepth); err != nil {
			return nil, err
		}
		data = append(data, analysis.UniqGroup{
			Name:           name,
			Exchange:       exchange,
			Symbol:         symbol,
			Unit:           domain.Unit(unit),
			Interval:       interval,
			Depth:          depth,
			ByIndicator:    byIndicator,
			IndicatorDepth: indicatorDepth,
		})
	}
	return data, nil
}

func (repo *AnalyticRepository) LastInGroup(ctx context.Context, g analysis.UniqGroup, offset int) (*analysis.Analytic, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_last_in_groups")()
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
WHERE name = ?
  AND exchange = ?
  AND symbol = ?
  AND unit = ?
  AND interval = ?
  AND depth = ?
  AND by_indicator = ?
  AND indicator_depth = ?
ORDER BY datetime DESC
LIMIT 1 OFFSET ?
`
	var data AnalyticDTO
	err := repo.db.GetContext(
		ctx,
		&data,
		query,
		g.Name, g.Exchange, g.Symbol, g.Unit, g.Interval, g.Depth, g.ByIndicator, g.IndicatorDepth,
		offset,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &analysis.Analytic{
		ID:             data.ID,
		Name:           data.Name,
		Exchange:       data.Exchange,
		Symbol:         data.Symbol,
		Unit:           domain.Unit(data.Unit),
		Interval:       data.Interval,
		Datetime:       data.Datetime,
		Depth:          data.Depth,
		ByIndicator:    data.ByIndicator,
		IndicatorDepth: data.IndicatorDepth,
		Value:          data.Value,
	}, nil
}

func (repo *AnalyticRepository) DeleteByGroup(ctx context.Context, g analysis.UniqGroup, datetime time.Time) error {
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_delete_by_group")()
	var query = `
DELETE
FROM analytics
WHERE name = ?
  AND exchange = ?
  AND symbol = ?
  AND unit = ?
  AND interval = ?
  AND depth = ?
  AND by_indicator = ?
  AND indicator_depth = ?
  AND datetime < ?;
`
	_, err := repo.db.ExecContext(
		ctx, query,
		g.Name, g.Exchange, g.Symbol, g.Unit, g.Interval, g.Depth, g.ByIndicator, g.IndicatorDepth, datetime,
	)

	return err
}

func (repo *AnalyticRepository) AllAnalyticInfo(ctx context.Context) (map[string][]view.AnalyticInfoModel, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "analysis_all_analytic_info")()
	result := make(map[string][]view.AnalyticInfoModel)
	var query = `SELECT 
    DISTINCT unit, interval, name, depth, indicator_depth
FROM analytics`
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			unit, name                      string
			interval, depth, indicatorDepth int
		)
		if err := rows.Scan(&unit, &interval, &name, &depth, &indicatorDepth); err != nil {
			return nil, err
		}
		model := view.AnalyticInfoModel{
			Unit:           unit,
			Interval:       interval,
			Name:           name,
			Depth:          depth,
			IndicatorDepth: indicatorDepth,
		}
		if _, has := result[model.Name]; !has {
			result[model.Name] = make([]view.AnalyticInfoModel, 0)
		}
		result[model.Name] = append(result[model.Name], model)
	}
	return result, nil
}
