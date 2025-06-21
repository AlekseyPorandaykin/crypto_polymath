package exchange

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
			Symbol:      resItem.Symbol,
			Exchange:    resItem.Exchange,
			BaseAsset:   resItem.BaseAsset,
			QuoteAsset:  resItem.QuoteAsset,
			IsExist:     true,
			FundingRate: resItem.FundingRate,
		}
		if !resItem.NextFundingTime.IsZero() {
			domainVal.NextFundingTime = &resItem.NextFundingTime
		}
		data = append(data, domainVal)
		storageDTOs = append(storageDTOs, SymbolInfoStorageDTO{
			ID:              uuid.New(),
			Exchange:        domainVal.Exchange,
			Symbol:          domainVal.Symbol,
			BaseAsset:       domainVal.BaseAsset,
			QuoteAsset:      domainVal.QuoteAsset,
			Category:        string(resItem.Category),
			FundingRate:     domainVal.FundingRate,
			NextFundingTime: domainVal.NextFundingTime,
			CreatedAt:       now,
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

func (s *service) SymbolInfoByCategory(ctx context.Context, exchange, category string) ([]domain.SymbolInfo, error) {
	data, err := s.repo.InfoByCategory(ctx, exchange, category)
	if err != nil {
		return nil, errors.Wrap(err, "fetch data from storage")
	}
	domains := make([]domain.SymbolInfo, 0, len(data))
	for _, item := range data {
		domains = append(domains, domain.SymbolInfo{
			Symbol:          item.Symbol,
			Exchange:        item.Exchange,
			BaseAsset:       item.BaseAsset,
			QuoteAsset:      item.QuoteAsset,
			IsExist:         true,
			FundingRate:     item.FundingRate,
			NextFundingTime: item.NextFundingTime,
		})
	}
	return domains, nil
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
