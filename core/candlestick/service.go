package candlestick

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/cenkalti/backoff/v4"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/pkg/errors"
	"time"
)

var ErrNotSupportInterval = errors.New("not support interval")

type service struct {
	loaders map[string]ExchangeLoader
	repo    Repository
}

func NewService(repo Repository) Candlestick {
	return &service{
		loaders: make(map[string]ExchangeLoader),
		repo:    repo,
	}
}

func (s *service) AddLoader(exchange string, loader ExchangeLoader) {
	s.loaders[exchange] = loader
}

func (s *service) LoadCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int) ([]domain.Candlestick, error) {
	data, err := s.loadCandlesticks(ctx, exchange, symbol, unit, interval)
	if err != nil {
		return nil, err
	}
	result, err := s.handleCandlesticks(ctx, data, unit, exchange, symbol, interval)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) loadCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int) ([]ExchangeDTO, error) {
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	switch unit {
	case domain.MinuteUnit:
		return loader.LastMinuteCandlesticks(ctx, symbol, interval)
	case domain.HourUnit:
		return loader.LastHourCandlesticks(ctx, symbol, interval)
	case domain.DayUnit:
		return loader.LastDayCandlesticks(ctx, symbol)
	case domain.WeekUnit:
		return loader.LastWeekCandlesticks(ctx, symbol)
	case domain.MonthUnit:
		return loader.LastMonthCandlesticks(ctx, symbol)
	default:
		return nil, nil
	}
}

func (s *service) SequenceCandlesticks(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int,
) ([]domain.Candlestick, error) {
	data, err := s.Candlesticks(ctx, exchange, symbol, unit, interval, limit)
	if err != nil {
		return nil, errors.Wrap(err, "get candlesticks")
	}
	if !domain.IsCorrectSequenceCandlesticks(data) {
		return []domain.Candlestick{}, nil
	}
	return data, nil
}

func (s *service) Candlesticks(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int,
) ([]domain.Candlestick, error) {
	switch unit {
	case domain.MinuteUnit:
		return s.candlesticksMinutes(ctx, exchange, symbol, interval, limit)
	case domain.HourUnit:
		return s.candlesticksHours(ctx, exchange, symbol, interval, limit)
	case domain.DayUnit:
		if interval != 1 {
			return nil, ErrNotSupportInterval
		}
		return s.candlesticksDay(ctx, exchange, symbol, limit)
	case domain.WeekUnit:
		if interval != 1 {
			return nil, ErrNotSupportInterval
		}
		return s.candlesticksWeek(ctx, exchange, symbol, limit)
	case domain.MonthUnit:
		if interval != 1 {
			return nil, ErrNotSupportInterval
		}
		return s.candlesticksMonth(ctx, exchange, symbol, limit)
	default:
		return nil, ErrNotSupportInterval
	}
}

func (s *service) candlesticksMinutes(
	ctx context.Context, exchange, symbol string, minutes, limit int,
) ([]domain.Candlestick, error) {
	if !isSupportMinute(minutes) {
		return nil, errors.New("don't support this minutes")
	}
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.MinuteUnit), minutes, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) candlesticksHours(
	ctx context.Context, exchange, symbol string, hours, limit int,
) ([]domain.Candlestick, error) {
	if !isSupportHours(hours) {
		return nil, errors.New("don't support this hours")
	}
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.HourUnit), hours, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) candlesticksDay(ctx context.Context, exchange, symbol string, limit int) ([]domain.Candlestick, error) {
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.DayUnit), 1, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) candlesticksWeek(ctx context.Context, exchange, symbol string, limit int) ([]domain.Candlestick, error) {
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.WeekUnit), 1, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) candlesticksMonth(ctx context.Context, exchange, symbol string, limit int) ([]domain.Candlestick, error) {
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.MonthUnit), 1, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) DeleteOldRows(ctx context.Context, oldValueLimit int) error {
	uniqData, err := s.repo.ListUniq(ctx)
	if err != nil {
		return err
	}
	for _, uniqItem := range uniqData {
		rows, err := s.repo.Last(
			ctx, uniqItem.Exchange, uniqItem.Symbol, uniqItem.Unit, uniqItem.Interval, 1, oldValueLimit,
		)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			continue
		}
		if err := s.repo.DeleteOldRows(
			ctx, uniqItem.Exchange, uniqItem.Symbol, uniqItem.Unit, uniqItem.Interval, rows[0].StartTime,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) UpdateCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int) error {
	data, err := s.loadCandlesticks(ctx, exchange, symbol, unit, interval)
	if err != nil {
		return err
	}
	if _, err := s.handleCandlesticks(ctx, data, unit, exchange, symbol, interval); err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	slice.SortBy[ExchangeDTO](data, func(a, b ExchangeDTO) bool {
		return a.StartTime.Before(b.StartTime)
	})
	if err := s.repo.DeleteOldRows(ctx, exchange, symbol, string(unit), interval, data[0].StartTime); err != nil {
		return err
	}
	return nil
}

func (s *service) CandlesticksToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]domain.Candlestick, error) {
	data, err := s.repo.LastToDate(ctx, exchange, symbol, unit, interval, limit, to)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Candlestick, 0, len(data))
	for _, item := range uniqStorageData(data) {
		result = append(result, storageToDomain(item))
	}
	return result, nil
}

func (s *service) SequenceCandlesticksToDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	data, err := s.CandlesticksToDate(ctx, exchange, symbol, unit, minutes, limit, to)
	if err != nil {
		return nil, err
	}
	if !domain.IsCorrectSequenceCandlesticks(data) {
		return nil, nil
	}
	return data, nil
}

func (s *service) CandlesticksFromDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	data, err := s.repo.FromDate(ctx, exchange, symbol, unit, minutes, limit, to)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Candlestick, 0, len(data))
	for _, item := range uniqStorageData(data) {
		result = append(result, storageToDomain(item))
	}
	return result, nil
}

func (s *service) candlesticks(
	ctx context.Context, exchange, symbol, unit string, minutes, limit int,
) ([]domain.Candlestick, error) {
	data, err := s.repo.Last(ctx, exchange, symbol, unit, minutes, limit, 0)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Candlestick, 0, len(data))
	for _, item := range uniqStorageData(data) {
		result = append(result, storageToDomain(item))
	}
	return result, nil
}

func (s *service) handleCandlesticks(
	ctx context.Context, data []ExchangeDTO, unit domain.Unit, exchange, symbol string, interval int,
) ([]domain.Candlestick, error) {
	lastRow := s.lastRow(ctx, exchange, symbol, string(unit), interval)
	storageData := make([]StorageDTO, 0, len(data))
	result := make([]domain.Candlestick, 0, len(data))
	slice.SortBy(data, func(a, b ExchangeDTO) bool {
		return a.StartTime.After(b.StartTime)
	})
	for i, item := range data {
		candle, err := exchangeToDomain(item, unit, exchange, symbol, interval)
		if err != nil {
			return nil, err
		}
		//Пропускаем если эта текущая свеча
		if domain.IsOpenCandle(candle) {
			continue
		}
		//Пропускаем если период еще не закрылся (первый в списке, новый период еще не начал расчет)
		if i == 0 {
			continue
		}
		if lastRow != nil && (candle.StartTime.Before(lastRow.StartTime) || candle.StartTime.Equal(lastRow.StartTime)) {
			continue
		}
		result = append(result, candle)
		storageItem, err := exchangeToStorage(item, unit, exchange, symbol, interval)
		if err != nil {
			return nil, err
		}
		storageData = append(storageData, storageItem)
	}
	if len(storageData) == 0 {
		return nil, nil
	}
	errSaver := backoff.Retry(func() error {
		return s.repo.Save(ctx, uniqStorageData(storageData)...)
	}, backoff.NewExponentialBackOff())
	if errSaver != nil {
		return nil, errSaver
	}
	slice.SortBy(result, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	return result, nil
}

func (s *service) lastRow(ctx context.Context, exchange, symbol, unit string, interval int) *StorageDTO {
	data, err := s.repo.Last(ctx, exchange, symbol, unit, interval, 1, 0)
	if err != nil {
		return nil
	}
	for _, item := range data {
		return &item
	}
	return nil
}

func uniqStorageData(data []StorageDTO) []StorageDTO {
	result := make([]StorageDTO, 0, len(data))
	existKey := make(map[time.Time]bool, len(data))
	for _, item := range data {
		if existKey[item.StartTime] {
			continue
		}
		result = append(result, item)
	}
	return result
}
