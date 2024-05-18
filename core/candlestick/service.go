package candlestick

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/cenkalti/backoff/v4"
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

func (s *service) LoadCandlesticksMinutes(
	ctx context.Context, exchange, symbol string, minutes int,
) ([]domain.Candlestick, error) {
	if !isSupportMinute(minutes) {
		return nil, errors.New("don't support this minutes")
	}
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	data, err := loader.LastMinuteCandlesticks(ctx, symbol, minutes)
	if err != nil {
		return nil, err
	}
	result, err := s.handleCandlesticks(ctx, data, domain.MinuteUnit, exchange, symbol, minutes)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) LoadCandlesticksHours(
	ctx context.Context, exchange, symbol string, hours int,
) ([]domain.Candlestick, error) {
	if !isSupportHours(hours) {
		return nil, errors.New("don't support this hours")
	}
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	data, err := loader.LastHourCandlesticks(ctx, symbol, hours)
	if err != nil {
		return nil, err
	}

	result, err := s.handleCandlesticks(ctx, data, domain.HourUnit, exchange, symbol, hours)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) LoadCandlesticksDay(ctx context.Context, exchange, symbol string) ([]domain.Candlestick, error) {
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	data, err := loader.LastDayCandlesticks(ctx, symbol)
	if err != nil {
		return nil, err
	}
	result, err := s.handleCandlesticks(ctx, data, domain.DayUnit, exchange, symbol, 1)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) LoadCandlesticksWeek(ctx context.Context, exchange, symbol string) ([]domain.Candlestick, error) {
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	data, err := loader.LastWeekCandlesticks(ctx, symbol)
	if err != nil {
		return nil, err
	}
	result, err := s.handleCandlesticks(ctx, data, domain.WeekUnit, exchange, symbol, 1)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) LoadCandlesticksMonth(ctx context.Context, exchange, symbol string) ([]domain.Candlestick, error) {
	loader, has := s.loaders[exchange]
	if !has || loader == nil {
		return nil, nil
	}
	data, err := loader.LastMonthCandlesticks(ctx, symbol)
	if err != nil {
		return nil, err
	}
	result, err := s.handleCandlesticks(ctx, data, domain.MonthUnit, exchange, symbol, 1)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) Save(ctx context.Context, candlesticks ...domain.Candlestick) error {
	data := make([]StorageDTO, 0, len(candlesticks))
	for _, candlestick := range candlesticks {
		data = append(data, domainToStorage(candlestick))
	}
	err := backoff.Retry(func() error {
		return s.repo.SaveBatch(ctx, data...)
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return errors.Wrap(err, "save candlestick")
	}
	return nil
}

func (s *service) Candlestick(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int,
) ([]domain.Candlestick, error) {
	switch unit {
	case domain.MinuteUnit:
		return s.CandlesticksMinutes(ctx, exchange, symbol, interval, limit)
	case domain.HourUnit:
		return s.CandlesticksHours(ctx, exchange, symbol, interval, limit)
	case domain.DayUnit:
		if interval != 1 {
			return nil, ErrNotSupportInterval
		}
		return s.CandlesticksDay(ctx, exchange, symbol, limit)
	case domain.WeekUnit:
		if interval != 1 {
			return nil, ErrNotSupportInterval
		}
		return s.CandlesticksWeek(ctx, exchange, symbol, limit)
	case domain.MonthUnit:
		if interval != 1 {
			return nil, ErrNotSupportInterval
		}
		return s.CandlesticksMonth(ctx, exchange, symbol, limit)
	default:
		return nil, ErrNotSupportInterval
	}
}
func (s *service) CandlesticksMinutes(
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

func (s *service) CandlesticksHours(
	ctx context.Context, exchange, symbol string, hours, limit int,
) ([]domain.Candlestick, error) {
	if !isSupportHours(hours) {
		return nil, errors.New("don't support this minutes")
	}
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.HourUnit), hours, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) CandlesticksDay(ctx context.Context, exchange, symbol string, limit int) ([]domain.Candlestick, error) {
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.DayUnit), 1, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) CandlesticksWeek(ctx context.Context, exchange, symbol string, limit int) ([]domain.Candlestick, error) {
	result, err := s.candlesticks(ctx, exchange, symbol, string(domain.WeekUnit), 1, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) CandlesticksMonth(ctx context.Context, exchange, symbol string, limit int) ([]domain.Candlestick, error) {
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

func (s *service) CandlesticksToDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	data, err := s.repo.LastToDate(ctx, exchange, symbol, unit, minutes, limit, to)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Candlestick, 0, len(data))
	for _, item := range data {
		result = append(result, storageToDomain(item))
	}
	return result, nil
}
func (s *service) CandlesticksFromDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	data, err := s.repo.FromDate(ctx, exchange, symbol, unit, minutes, limit, to)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Candlestick, 0, len(data))
	for _, item := range data {
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
	for _, item := range data {
		result = append(result, storageToDomain(item))
	}
	return result, nil
}

func (s *service) handleCandlesticks(
	ctx context.Context, data []ExchangeDTO, unit domain.Unit, exchange, symbol string, interval int,
) ([]domain.Candlestick, error) {
	lastRow := s.lastRow(ctx, exchange, symbol, string(unit), interval)
	now := time.Now().In(time.UTC)
	storageData := make([]StorageDTO, 0, len(data))
	result := make([]domain.Candlestick, 0, len(data))
	for i, item := range data {
		if unit == domain.MinuteUnit && interval == 1 && now.Minute() != item.StartTime.Minute() && i == 0 {
			continue
		}
		candle, err := exchangeToDomain(item, unit, exchange, symbol, interval)
		if err != nil {
			return nil, err
		}
		if ignoreOpenCandle(now, candle) {
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
		return s.repo.SaveBatch(ctx, storageData...)
	}, backoff.NewExponentialBackOff())
	if errSaver != nil {
		return nil, errSaver
	}
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

// TODO Доделать фильтрацию
func ignoreOpenCandle(now time.Time, candlestick domain.Candlestick) bool {
	switch candlestick.Unit {
	case domain.MonthUnit:
		return candlestick.StartTime.Year() == now.Year() && candlestick.StartTime.Month() == now.Month()
	case domain.WeekUnit:
		return candlestick.StartTime.Year() == now.Year() &&
			candlestick.StartTime.Month() == now.Month() &&
			candlestick.StartTime.Weekday() == now.Weekday()
	case domain.DayUnit:
		return candlestick.StartTime.Year() == now.Year() &&
			candlestick.StartTime.Month() == now.Month() &&
			candlestick.StartTime.Day() == now.Day()
	case domain.HourUnit:
		return candlestick.StartTime.Year() == now.Year() &&
			candlestick.StartTime.Month() == now.Month() &&
			candlestick.StartTime.Day() == now.Day() &&
			candlestick.StartTime.Hour() == now.Hour()
	case domain.MinuteUnit:
		return candlestick.StartTime.Year() == now.Year() &&
			candlestick.StartTime.Month() == now.Month() &&
			candlestick.StartTime.Day() == now.Day() &&
			candlestick.StartTime.Hour() == now.Hour() &&
			candlestick.StartTime.Minute() == now.Minute()
	}
	return false
}
