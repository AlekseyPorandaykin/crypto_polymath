package indicator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type service struct {
	calculators map[string]Calculator
	repo        Repository
	candlestick Candlestick
}

func NewService(repo Repository, candlestick Candlestick) Indicator {
	s := &service{
		repo:        repo,
		candlestick: candlestick,
		calculators: make(map[string]Calculator),
	}
	s.calculators[domain.MAIndicator] = NewMA()
	s.calculators[domain.EMAIndicator] = NewEMA()
	s.calculators[domain.TypeCandleIndicator] = NewTypeCandle()
	s.calculators[domain.VolatilityCandlePercentIndicator] = NewVolatilityCandlePercent()
	s.calculators[domain.TrendIndicator] = NewTrend()
	s.calculators[domain.PriceChanges] = NewPriceChanges()
	return s
}

func (s *service) AddCalculator(calculator Calculator) {
	s.calculators[calculator.Name()] = calculator
}

func (s *service) CalcIndicators(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, depth int) (int, error) {
	var totalCount int
	for key := range s.calculators {
		count, err := s.CalculateLastCandles(ctx, exchange, symbol, unit, interval, key, depth)
		if err != nil {
			return count, err
		}
		totalCount += count
	}
	return totalCount, nil
}

func (s *service) Indicator(ctx context.Context, candlestick domain.Candlestick, name string, depth int) (*domain.Indicator, error) {
	storageDTO, err := s.repo.Find(
		ctx,
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
		res := storageToDomain(*storageDTO)
		return &res, nil
	}

	return nil, nil
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
func (s *service) CalculateIndicator(ctx context.Context, candlestick domain.Candlestick, name string, depth int) (*domain.Indicator, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
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
	calc, has := s.calculators[name]
	if !has || calc == nil {
		return nil, nil
	}
	if depth < 1 {
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

func (s *service) CalculateLastCandles(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, indicator string, depth int,
) (int, error) {
	i := 0
	for {
		calc, has := s.calculators[indicator]
		if !has || calc == nil {
			return i, nil
		}
		if !calc.SupportInterval(interval) || !calc.SupportDepth(depth) {
			return i, nil
		}
		lastCandle, err := s.candleForCalculate(ctx, exchange, symbol, unit, interval, indicator, depth)
		if err != nil {
			return i, err
		}
		if lastCandle == nil {
			return i, nil
		}
		res, err := s.CalculateIndicator(ctx, *lastCandle, indicator, depth)
		if err != nil {
			return i, err
		}
		if res == nil {
			return i, nil
		}
		i++
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
