package exchange

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"strings"
	"time"
)

type service struct {
	loaders map[string]ExternalLoader
	repo    Repository
}

func New(repo Repository) Exchange {
	return &service{
		loaders: make(map[string]ExternalLoader),
		repo:    repo,
	}
}

func (s *service) AddLoader(exchangeName string, loader ExternalLoader) {
	s.loaders[exchangeName] = loader
}

func (s *service) LoadSymbolInfo(ctx context.Context, exchange string) ([]domain.SymbolInfo, error) {
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	res, err := loader.SymbolInfo(ctx)
	if err != nil {
		return nil, err
	}
	data := make([]domain.SymbolInfo, 0, len(res))
	storageDTOs := make([]SymbolInfoStorageDTO, 0, len(res))
	now := time.Now().In(time.UTC)
	for _, resItem := range res {
		domainVal := domain.SymbolInfo{
			Symbol:     resItem.Symbol,
			Exchange:   resItem.Exchange,
			BaseAsset:  resItem.BaseAsset,
			QuoteAsset: resItem.QuoteAsset,
			IsExist:    true,
		}
		data = append(data, domainVal)
		storageDTOs = append(storageDTOs, SymbolInfoStorageDTO{
			ID:         uuid.New(),
			Exchange:   resItem.Exchange,
			Symbol:     resItem.Symbol,
			BaseAsset:  resItem.BaseAsset,
			QuoteAsset: resItem.QuoteAsset,
			Category:   string(resItem.Category),
			CreatedAt:  now,
		})
	}
	if err := s.repo.SaveSymbolInfo(ctx, storageDTOs); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteOldRows(ctx, exchange, now); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *service) SymbolInfo(ctx context.Context, exchange, symbol string) (*domain.SymbolInfo, error) {
	data, err := s.repo.InfoBySymbol(ctx, exchange, symbol)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return &domain.SymbolInfo{
			Symbol:     data.Symbol,
			Exchange:   data.Exchange,
			BaseAsset:  data.BaseAsset,
			QuoteAsset: data.QuoteAsset,
			IsExist:    true,
		}, nil
	}
	//Символ уже мог быть удален, попробуем найти его по котируемому активу
	quoteAssets, err := s.repo.QuoteAssets(ctx)
	if err != nil {
		return nil, err
	}
	for _, quoteAsset := range quoteAssets {
		if !strings.HasSuffix(symbol, quoteAsset) {
			continue
		}
		baseAsset := strings.TrimPrefix(symbol, quoteAsset)
		return &domain.SymbolInfo{
			Symbol:     symbol,
			Exchange:   exchange,
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
			IsExist:    false,
		}, nil
	}
	return nil, nil
}
