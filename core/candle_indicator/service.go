package candle_indicator

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/pkg/errors"
)

type CandleIndicator interface {
	AddCalculator(c Calculator)
	CalculateFromCandlesticks(ctx context.Context, data []domain.Candlestick) ([]Indicator, error)
	Indicators(ctx context.Context, name, exchange, symbol string, unit domain.Unit, interval int) ([]Indicator, error)
}

type Service struct {
	repo               Repository
	candlestickService candlestick.Candlestick
	calculators        map[string]Calculator
}

func New(repo Repository, candlestickService candlestick.Candlestick) CandleIndicator {
	return &Service{
		repo:               repo,
		candlestickService: candlestickService,
		calculators:        make(map[string]Calculator),
	}
}

func (s *Service) AddCalculator(c Calculator) {
	s.calculators[c.Name()] = c
}

func (s *Service) Indicators(ctx context.Context, name, exchange, symbol string, unit domain.Unit, interval int) ([]Indicator, error) {
	data, err := s.repo.FetchLast(ctx, name, exchange, symbol, string(unit), interval)
	if err != nil {
		return nil, errors.Wrap(err, "fetch last indicators")
	}
	if len(data) < 100 {
		return s.calculateLastCandlesticks(ctx, name, exchange, symbol, unit, interval)
	}
	result := make([]Indicator, 0, len(data))
	for _, item := range data {
		result = append(result, StorageToDomain(item))
	}
	return result, nil
}

func (s *Service) calculateLastCandlesticks(
	ctx context.Context, name, exchange, symbol string, unit domain.Unit, interval int,
) ([]Indicator, error) {
	data, err := s.candlestickService.Candlesticks(ctx, exchange, symbol, unit, interval, 100)
	if err != nil {
		return nil, errors.Wrap(err, "get candlesticks")
	}
	return s.CalculateFromCandlesticks(ctx, data)
}

func (s *Service) CalculateFromCandlesticks(ctx context.Context, data []domain.Candlestick) ([]Indicator, error) {
	if len(data) == 0 {
		return nil, nil
	}
	grouped := make(map[string][]domain.Candlestick)
	for _, item := range data {
		key := fmt.Sprintf("%s_%s_%s_%d", item.Exchange, item.Symbol, item.Unit, item.Interval)
		if _, has := grouped[key]; !has {
			grouped[key] = make([]domain.Candlestick, 0, len(data))
		}
		grouped[key] = append(grouped[key], item)
	}
	result := make([]Indicator, 0, len(data))
	for _, candles := range grouped {
		slice.SortBy[domain.Candlestick](candles, func(a, b domain.Candlestick) bool {
			return a.StartTime.Before(b.StartTime)
		})
		for _, candle := range candles {
			for key, c := range s.calculators {
				indicatorStorage, err := s.repo.Find(
					ctx, c.Name(), candle.Exchange, candle.Symbol, string(candle.Unit), candle.Interval, candle.StartTime,
				)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("find indicator %s", key))
				}
				if indicatorStorage != nil {
					result = append(result, StorageToDomain(*indicatorStorage))
					continue
				}
				indicator, err := c.Calculate(ctx, candle)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("calculate indicator %s", key))
				}
				if indicator == nil {
					continue
				}
				result = append(result, *indicator)
				if err := s.repo.Save(ctx, []StorageDTO{DomainToStorage(*indicator)}); err != nil {
					return nil, errors.Wrap(err, "save indicator")
				}
			}
		}
	}

	return result, nil
}
