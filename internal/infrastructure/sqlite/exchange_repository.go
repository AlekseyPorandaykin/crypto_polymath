package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type ExchangeRepository struct {
	db *sqlx.DB
}

func NewExchangeRepository(db *sqlx.DB) *ExchangeRepository {
	return &ExchangeRepository{db: db}
}

func (repo *ExchangeRepository) SaveSymbolInfo(ctx context.Context, data []core_exchange.SymbolInfoStorageDTO) error {
	defer metrics.DBQueryHelper("crypto_polymath", "exchange_save_symbol_info")()
	query := `
INSERT INTO symbol_infos (id, exchange, symbol, base_asset, quote_asset, category, created_at)
VALUES 
`
	valueQuery := `(:id, :exchange, :symbol, :base_asset, :quote_asset, :category, :created_at)`
	values := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)*6)
	for _, item := range data {
		preparedValues, argsItem, err := repo.db.BindNamed(valueQuery, item)
		if err != nil {
			return nil
		}
		values = append(values, preparedValues)
		args = append(args, argsItem...)
	}
	preparedQuery := fmt.Sprintf("%s %s", query, strings.Join(values, ", "))
	if _, err := repo.db.ExecContext(ctx, preparedQuery, args...); err != nil {
		return err
	}
	return nil
}

func (repo *ExchangeRepository) InfoBySymbol(ctx context.Context, exchange, symbol string) (*core_exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "exchange_fetch_info_by_symbol")()
	var (
		query = `
SELECT id, exchange, symbol, base_asset, quote_asset, created_at
FROM symbol_infos
WHERE exchange= ? AND symbol = ?
LIMIT 1
`
		result core_exchange.SymbolInfoStorageDTO
	)
	err := repo.db.GetContext(ctx, &result, query, exchange, symbol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (repo *ExchangeRepository) InfoByCategory(ctx context.Context, exchange, category string) ([]core_exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "exchange_fetch_infos_by_category")()
	var (
		query = `
SELECT id, exchange, symbol, base_asset, quote_asset, created_at
FROM symbol_infos
WHERE exchange = ? AND category = ?
ORDER BY symbol
`
		result []core_exchange.SymbolInfoStorageDTO
	)
	err := repo.db.SelectContext(ctx, &result, query, exchange, category)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *ExchangeRepository) QuoteAssets(ctx context.Context) ([]string, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "exchange_quote_assets")()
	query := `
SELECT DISTINCT quote_asset
FROM symbol_infos
`
	quotes := make([]string, 0, 100)
	if err := repo.db.SelectContext(ctx, &quotes, query); err != nil {
		return nil, err
	}
	return quotes, nil
}

func (repo *ExchangeRepository) DeleteOldRows(ctx context.Context, exchangeName string, to time.Time) error {
	defer metrics.DBQueryHelper("crypto_polymath", "exchange_delete_old_rows")()
	var query = `DELETE FROM symbol_infos WHERE exchange=? AND created_at < ?`
	if _, err := repo.db.ExecContext(ctx, query, exchangeName, to); err != nil {
		return err
	}
	return nil
}

func (repo *ExchangeRepository) ListByExchange(ctx context.Context, exchangeName string) ([]core_exchange.SymbolInfoStorageDTO, error) {
	var query = `
SELECT id, exchange, symbol, base_asset, quote_asset, created_at
FROM symbol_infos
WHERE exchange = ?
`
	list := make([]core_exchange.SymbolInfoStorageDTO, 0, 1_000)
	if err := repo.db.SelectContext(ctx, &list, query, exchangeName); err != nil {
		return nil, err
	}
	return list, nil
}
