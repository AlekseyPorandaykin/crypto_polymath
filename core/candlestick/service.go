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

func (s *service) loadCandlesticksMinutes(
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

func (s *service) Save(ctx context.Context, candlesticks ...domain.Candlestick) error {
	data := make([]StorageDTO, 0, len(candlesticks))
	for _, candlestick := range candlesticks {
		data = append(data, domainToStorage(candlestick))
	}
	err := backoff.Retry(func() error {
		return s.repo.Save(ctx, data...)
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
		return nil, errors.New("don't support this minutes")
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

func (s *service) CandlesticksToDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	data, err := s.repo.LastToDate(ctx, exchange, symbol, unit, minutes, limit, to)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Candlestick, 0, len(data))
	for _, item := range uniqStorageData(data) {
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
	for i, item := range data {
		candle, err := exchangeToDomain(item, unit, exchange, symbol, interval)
		if err != nil {
			return nil, err
		}
		//Пропускаем если эта текущая свеча
		if isOpenCandle(candle) {
			continue
		}
		//Пропускаем если период еще не закрылся (первый в списке, новый период еще не начал рассчет)
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

// TODO Доделать фильтрацию
func isOpenCandle(candlestick domain.Candlestick) bool {
	now := time.Now().In(time.UTC)
	startTime := candlestick.StartTime.In(time.UTC)
	switch candlestick.Unit {
	case domain.MonthUnit:
		return startTime.Year() == now.Year() && startTime.Month() == now.Month()
	case domain.WeekUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Weekday() == now.Weekday()
	case domain.DayUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day()
	case domain.HourUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day() &&
			startTime.Hour() == now.Hour()
	case domain.MinuteUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day() &&
			startTime.Hour() == now.Hour() &&
			startTime.Minute() == now.Minute()
	}
	return false
}
