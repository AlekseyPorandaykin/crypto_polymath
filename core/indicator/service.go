package indicator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type service struct {
	calculators map[string]calculator.PrimaryIndicatorCalculator
	repo        Repository
	candlestick Candlestick
}

func NewService(repo Repository, candlestick Candlestick) Indicator {
	return &service{
		repo:        repo,
		candlestick: candlestick,
		calculators: make(map[string]calculator.PrimaryIndicatorCalculator),
	}
}

func (s *service) AddCalculator(calculator calculator.PrimaryIndicatorCalculator) {
	s.calculators[calculator.Name()] = calculator
}

func (s *service) CalcIndicators(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, depth int) ([]domain.Indicator, error) {
	indicators := make([]domain.Indicator, 0, 100)
	for key := range s.calculators {
		indicatorsLastCandle, err := s.calculateLastCandles(ctx, exchange, symbol, unit, interval, key, depth)
		if err != nil {
			return indicators, err
		}
		indicators = append(indicators, indicatorsLastCandle...)
	}
	return indicators, nil
}

func (s *service) CalcIndicatorsByCandlestick(ctx context.Context, candlestick domain.Candlestick, depth int) ([]domain.Indicator, error) {
	indicators := make([]domain.Indicator, 0, len(s.calculators))
	for key := range s.calculators {
		primaryIndicator, err := s.CalculatePrimaryIndicator(ctx, candlestick, key, depth)
		if err != nil {
			return nil, err
		}
		if primaryIndicator == nil {
			continue
		}
		indicators = append(indicators, *primaryIndicator)
	}
	return indicators, nil
}

func (s *service) Indicators(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, name string, depth, limit int,
) ([]domain.Indicator, error) {
	data, err := s.repo.List(ctx, exchange, symbol, string(unit), interval, name, depth, limit, 0)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Indicator, 0, len(data))
	for _, item := range data {
		res = append(res, storageToDomain(item))
	}

	return res, nil
}
func (s *service) CalculatePrimaryIndicator(ctx context.Context, candlestick domain.Candlestick, name string, depth int) (*domain.Indicator, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	calc, has := s.calculators[name]
	if !has || calc == nil {
		return nil, nil
	}
	if depth < 1 {
		return nil, nil
	}
	storageDTO, err := s.repo.Find(
		childCtx,
		candlestick.Exchange,
		candlestick.Symbol,
		string(candlestick.Unit),
		candlestick.Interval,
		candlestick.StartTime,
		name,
		depth,
	)
	if err != nil {
		return nil, errors.Wrap(err, "find indicator in storage")
	}
	if storageDTO != nil {
		return nil, nil
	}

	data, err := s.candlestick.LastCandlesticks(
		childCtx, candlestick.Exchange, candlestick.Symbol, candlestick.Unit, candlestick.Interval, depth, candlestick.StartTime,
	)
	if err != nil {
		return nil, err
	}
	if len(data) != depth {
		return nil, nil
	}
	res := calc.Calculate(data)

	if res == nil {
		return nil, nil
	}
	if err := s.repo.Save(childCtx, domainToStorage(*res)); err != nil {
		return nil, err
	}

	return res, nil
}

// TODO: надо возвращать последние записи, которые идут друг за другом или ничего
// иначе можно получить несколько записей, между которыми большие разрывы
func (s *service) LastToDate(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, name string, depth, limit int, to time.Time) ([]domain.Indicator, error) {
	data, err := s.repo.LastToDate(ctx, exchange, symbol, string(unit), interval, name, depth, limit, to)
	if err != nil {
		return nil, err
	}
	return storageToDomains(data), nil
}

func (s *service) calculateLastCandles(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, indicator string, depth int,
) ([]domain.Indicator, error) {
	indicators := make([]domain.Indicator, 0, 100)
	for {
		calc, has := s.calculators[indicator]
		if !has || calc == nil {
			return indicators, nil
		}
		if !calc.SupportInterval(interval) || !calc.SupportDepth(depth) {
			return indicators, nil
		}
		lastCandle, err := s.candleForCalculate(ctx, exchange, symbol, unit, interval, indicator, depth)
		if err != nil {
			return indicators, err
		}
		if lastCandle == nil {
			return indicators, nil
		}
		res, err := s.CalculatePrimaryIndicator(ctx, *lastCandle, indicator, depth)
		if err != nil {
			return nil, err
		}
		if res == nil {
			return indicators, nil
		}
		indicators = append(indicators, *res)
	}

}
func (s *service) candleForCalculate(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, indicator string, depth int,
) (*domain.Candlestick, error) {
	lastRow, err := s.repo.Last(ctx, exchange, symbol, string(unit), interval, indicator, depth)
	if err != nil {
		return nil, err
	}
	if lastRow != nil {
		nextIntervalVal := nextInterval(lastRow.Datetime, unit, interval)
		if isNextIntervalInPassed(nextIntervalVal, unit, interval) {
			nextCandles, err := s.nextCandle(ctx, exchange, symbol, unit, interval, nextIntervalVal)
			if err != nil {
				return nil, err
			}
			if nextCandles != nil {
				return nextCandles, nil
			}
		}
	}
	firstCandle, err := s.candlestick.FirstCandlestick(ctx, exchange, symbol, unit, interval, depth)
	if err != nil {
		return nil, err
	}

	return firstCandle, nil
}

func (s *service) nextCandle(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, nextInterval time.Time) (*domain.Candlestick, error) {
	nextCandlesticks, err := s.candlestick.NextCandlesticks(ctx, exchange, symbol, unit, interval, 10, nextInterval)
	if err != nil {
		return nil, err
	}
	if len(nextCandlesticks) == 0 {
		return nil, nil
	}
	slice.SortBy[domain.Candlestick](nextCandlesticks, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	return &nextCandlesticks[0], nil
}

func nextInterval(from time.Time, unit domain.Unit, interval int) time.Time {
	switch unit {
	case domain.MinuteUnit:
		return from.Add(time.Duration(interval) * time.Minute)
	case domain.HourUnit:
		return from.Add(time.Duration(interval) * time.Hour)
	case domain.DayUnit:
		return from.Add(time.Duration(interval) * 24 * time.Hour)
	case domain.WeekUnit:
		return from.Add(time.Duration(interval) * 7 * 24 * time.Hour)
	case domain.MonthUnit:
		return from.Add(time.Duration(interval) * 30 * 24 * time.Hour)
	default:
		return from
	}
}

func (s *service) DeleteOldRows(ctx context.Context, oldValueThreshold int) error {
	uniqData, err := s.repo.ListUniq(ctx)
	if err != nil {
		return errors.Wrap(err, "get list uniq indicators")
	}
	for _, uniqItem := range uniqData {
		row, err := s.repo.List(
			ctx,
			uniqItem.Exchange,
			uniqItem.Symbol,
			uniqItem.Unit,
			uniqItem.Interval,
			uniqItem.Name,
			uniqItem.Depth,
			1,
			oldValueThreshold,
		)
		if len(row) == 0 {
			continue
		}
		if err != nil {
			return err
		}
		errDelete := s.repo.DeleteOldRows(
			ctx,
			uniqItem.Symbol,
			uniqItem.Exchange,
			uniqItem.Unit,
			uniqItem.Interval,
			uniqItem.Name,
			uniqItem.Depth,
			row[0].Datetime,
		)
		if errDelete != nil {
			return errors.Wrap(errDelete, "delete old indicators")
		}

	}
	return nil
}

// Следующий интервал уже начат, т.е. текущий интервал уже закрыт
func isNextIntervalInPassed(nextIntervalVal time.Time, unit domain.Unit, interval int) bool {
	now := time.Now().In(time.UTC)
	closedInterval := nextInterval(nextIntervalVal, unit, interval) //Когда последний интервал должен появиться
	return now.After(closedInterval.In(time.UTC))
}

func storageToDomains(data []StorageDTO) []domain.Indicator {
	indicators := make([]domain.Indicator, 0, len(data))
	for _, item := range data {
		indicators = append(indicators, storageToDomain(item))
	}

	return indicators
}

func storageToDomain(data StorageDTO) domain.Indicator {
	return domain.Indicator{
		Symbol:   data.Symbol,
		Exchange: data.Exchange,
		Unit:     domain.Unit(data.Unit),
		Interval: data.Interval,
		Name:     data.Name,
		Datetime: data.Datetime.In(time.UTC),
		Depth:    data.Depth,
		Value:    data.Value,
	}
}

func domainToStorage(data domain.Indicator) StorageDTO {
	return StorageDTO{
		ID:        uuid.New(),
		Symbol:    data.Symbol,
		Exchange:  data.Exchange,
		Unit:      string(data.Unit),
		Interval:  data.Interval,
		Name:      data.Name,
		Datetime:  data.Datetime.In(time.UTC),
		Depth:     data.Depth,
		Value:     data.Value,
		CreatedAt: time.Now().In(time.UTC),
	}
}
