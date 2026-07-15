package analysis

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/pkg/errors"
)

// Service - сервисный слой (приложением), реализует логику работы с аналитикой
type Service struct {
	calculatorsByIndicator map[string]CalculatorByIndicator
	calculatorsByAnalytic  map[string]CalculatorByAnalytic
	repo                   Repository
	indicatorService       indicator.Indicator
	depths                 []int
}

func NewService(repo Repository, indicatorService indicator.Indicator, depths []int) *Service {
	return &Service{
		calculatorsByIndicator: make(map[string]CalculatorByIndicator),
		calculatorsByAnalytic:  make(map[string]CalculatorByAnalytic),
		repo:                   repo,
		indicatorService:       indicatorService,
		depths:                 depths,
	}
}

func (s *Service) AddCalculatorByIndicator(calc CalculatorByIndicator) {
	s.calculatorsByIndicator[calc.Name()] = calc
}

func (s *Service) AddCalculatorByAnalytic(calc CalculatorByAnalytic) {
	s.calculatorsByAnalytic[calc.Name()] = calc
}

func (s *Service) CalculateByIndicator(ctx context.Context, indicator domain.Indicator) ([]Analytic, error) {
	if len(s.calculatorsByIndicator) == 0 || len(s.depths) == 0 {
		return nil, nil
	}
	names := make([]string, 0, len(s.calculatorsByIndicator))
	for name := range s.calculatorsByIndicator {
		names = append(names, name)
	}
	existing, err := s.repo.FindManyByIndicator(
		ctx, indicator.Exchange, indicator.Symbol, indicator.Unit, indicator.Interval, indicator.Datetime, indicator.Depth, names, s.depths,
	)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	existingByNameDepth := make(map[string]map[int]Analytic, len(names))
	for _, item := range existing {
		if existingByNameDepth[item.Name] == nil {
			existingByNameDepth[item.Name] = make(map[int]Analytic, len(s.depths))
		}
		existingByNameDepth[item.Name][item.Depth] = item
	}

	calculatedAnalytics := make([]Analytic, 0, len(names)*len(s.depths))
	toSave := make([]Analytic, 0, len(names)*len(s.depths))
	for analyticName, calc := range s.calculatorsByIndicator {
		for _, depth := range s.depths {
			if data, ok := existingByNameDepth[analyticName][depth]; ok {
				calculatedAnalytics = append(calculatedAnalytics, data)
				continue
			}
			analytic, errCalc := s.computeAnalyticByIndicator(ctx, calc, indicator, depth)
			if errCalc != nil {
				return nil, errCalc
			}
			if analytic == nil {
				continue
			}
			calculatedAnalytics = append(calculatedAnalytics, *analytic)
			toSave = append(toSave, *analytic)
		}
	}
	if len(toSave) > 0 {
		if err := s.repo.Save(ctx, toSave...); err != nil {
			return nil, errors.Wrap(err, "save analytic")
		}
	}
	return calculatedAnalytics, nil
}

func (s *Service) AnalyticByIndicators(ctx context.Context, indicators []domain.Indicator, analyticName string, depth int) ([]Analytic, error) {
	if len(indicators) == 0 {
		return nil, nil
	}
	calc, has := s.calculatorsByIndicator[analyticName]
	if !has {
		return nil, nil
	}
	first := indicators[0]
	datetimes := make([]time.Time, len(indicators))
	for i, item := range indicators {
		datetimes[i] = item.Datetime
	}
	existing, err := s.repo.FindMany(
		ctx, analyticName, first.Exchange, first.Symbol, first.Unit, first.Interval, first.Depth, depth, datetimes,
	)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	existingByTime := make(map[time.Time]Analytic, len(existing))
	for _, item := range existing {
		existingByTime[item.Datetime] = item
	}

	result := make([]Analytic, 0, len(indicators))
	toSave := make([]Analytic, 0, len(indicators))
	for _, indicatorItem := range indicators {
		if data, ok := existingByTime[indicatorItem.Datetime]; ok {
			result = append(result, data)
			continue
		}
		analytic, errCalc := s.computeAnalyticByIndicator(ctx, calc, indicatorItem, depth)
		if errCalc != nil {
			return nil, errCalc
		}
		if analytic == nil {
			continue
		}
		result = append(result, *analytic)
		toSave = append(toSave, *analytic)
	}
	if len(toSave) > 0 {
		if err := s.repo.Save(ctx, toSave...); err != nil {
			return nil, errors.Wrap(err, "save analytic")
		}
	}
	return result, nil
}

// computeAnalyticByIndicator - Считает значение аналитики по индикатору, без обращения к storage за поиском/сохранением.
func (s *Service) computeAnalyticByIndicator(ctx context.Context, calc CalculatorByIndicator, indicator domain.Indicator, depth int) (*Analytic, error) {
	if indicator.Name != calc.ByIndicator() {
		return nil, nil
	}
	if !calc.SupportInterval(indicator.Interval) {
		return nil, nil
	}
	if !calc.SupportDepth(depth) {
		return nil, nil
	}
	return calc.Calculate(ctx, indicator, depth)
}

func (s *Service) CalculateByAnalytic(ctx context.Context, data Analytic) ([]Analytic, error) {
	return s.CalculateByAnalytics(ctx, []Analytic{data})
}

// CalculateByAnalytics — пакетный расчёт производных аналитик (MACD Signal Line, Histogram и т.д.).
// Один FindMany на пачку datetime вместо отдельного запроса на каждое сообщение.
func (s *Service) CalculateByAnalytics(ctx context.Context, data []Analytic) ([]Analytic, error) {
	if len(data) == 0 || len(s.calculatorsByAnalytic) == 0 || len(s.depths) == 0 {
		return nil, nil
	}
	bySourceName := make(map[string][]Analytic, len(data))
	for _, item := range data {
		bySourceName[item.Name] = append(bySourceName[item.Name], item)
	}

	calculatedAnalytics := make([]Analytic, 0, len(data)*len(s.calculatorsByAnalytic))
	for _, calc := range s.calculatorsByAnalytic {
		sourceAnalytics := bySourceName[calc.ByAnalytic()]
		if len(sourceAnalytics) == 0 {
			continue
		}
		for _, depth := range s.depths {
			if !calc.SupportDepth(depth) {
				continue
			}
			analytics, err := s.OscillatorByAnalytics(ctx, sourceAnalytics, calc.Name(), depth)
			if err != nil {
				return nil, err
			}
			calculatedAnalytics = append(calculatedAnalytics, analytics...)
		}
	}
	return calculatedAnalytics, nil
}

func (s *Service) OscillatorByAnalytics(ctx context.Context, data []Analytic, oscillatorName string, depth int) ([]Analytic, error) {
	if len(data) == 0 {
		return nil, nil
	}
	calc, has := s.calculatorsByAnalytic[oscillatorName]
	if !has {
		return nil, nil
	}
	first := data[0]
	datetimes := make([]time.Time, len(data))
	for i, item := range data {
		datetimes[i] = item.Datetime
	}
	existing, err := s.repo.FindMany(
		ctx, oscillatorName, first.Exchange, first.Symbol, first.Unit, first.Interval, first.Depth, depth, datetimes,
	)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	existingByTime := make(map[time.Time]Analytic, len(existing))
	for _, item := range existing {
		existingByTime[item.Datetime] = item
	}

	result := make([]Analytic, 0, len(data))
	toSave := make([]Analytic, 0, len(data))
	for _, analyticItem := range data {
		if found, ok := existingByTime[analyticItem.Datetime]; ok {
			result = append(result, found)
			continue
		}
		oscillator, errCalc := s.computeOscillatorByAnalytic(ctx, calc, analyticItem, oscillatorName, depth)
		if errCalc != nil {
			return nil, errCalc
		}
		if oscillator == nil {
			continue
		}
		result = append(result, *oscillator)
		toSave = append(toSave, *oscillator)
	}
	if len(toSave) > 0 {
		if err := s.repo.Save(ctx, toSave...); err != nil {
			return nil, errors.Wrap(err, "save oscillatorByAnalytic")
		}
	}
	return result, nil
}

// computeOscillatorByAnalytic - Считает значение осциллятора по аналитике, без обращения к storage за поиском/сохранением.
func (s *Service) computeOscillatorByAnalytic(ctx context.Context, calc CalculatorByAnalytic, data Analytic, oscillatorName string, depth int) (*Analytic, error) {
	if data.Name != calc.ByAnalytic() {
		return nil, nil
	}
	if !calc.SupportInterval(data.Interval) {
		return nil, nil
	}
	if calc.Name() != oscillatorName {
		return nil, nil
	}
	if !calc.SupportDepth(depth) {
		return nil, nil
	}
	return calc.Calculate(ctx, data, depth)
}

func (s *Service) LastAnalytics(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int,
) ([]Analytic, error) {
	dataByIndicator, err := s.calculateAnalyticByIndicator(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	if len(dataByIndicator) > 0 {
		return dataByIndicator, nil
	}

	oscillatorByIndicator, errOscillatorByIndicator := s.calculateOscillatorByIndicator(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if errOscillatorByIndicator != nil {
		return nil, errors.Wrap(errOscillatorByIndicator, "calculate oscillator by indicator")
	}
	if len(oscillatorByIndicator) > 0 {
		return oscillatorByIndicator, nil
	}
	return s.calculateOscillatorByAnalytics(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
}

func (s *Service) AnalyticsToDate(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int, datetime time.Time,
) ([]Analytic, error) {
	dataByIndicator, err := s.calculateAnalyticByIndicatorToDate(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth, datetime)
	if err != nil {
		return nil, err
	}
	if len(dataByIndicator) > 0 {
		return dataByIndicator, nil
	}

	oscillatorByIndicator, errOscillatorByIndicator := s.calculateAnalyticByIndicatorToDate(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth, datetime)
	if errOscillatorByIndicator != nil {
		return nil, errors.Wrap(errOscillatorByIndicator, "calculate oscillator by indicator")
	}
	if len(oscillatorByIndicator) > 0 {
		return oscillatorByIndicator, nil
	}
	return s.calculateOscillatorByIndicatorToDate(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth, datetime)
}

func (s *Service) calculateOscillatorByAnalytics(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int) ([]Analytic, error) {
	calc, has := s.calculatorsByAnalytic[name]
	if !has {
		return nil, nil
	}
	sourceAnalytics, err := s.calculateOscillatorByIndicator(ctx, exchangeName, symbol, unit, interval, calc.ByAnalytic(), indicatorDepth, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch source analytics")
	}
	return s.OscillatorByAnalytics(ctx, sourceAnalytics, name, depth)
}

func (s *Service) calculateOscillatorByIndicator(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int) ([]Analytic, error) {
	calc, has := s.calculatorsByAnalytic[name]
	if !has {
		return nil, nil
	}
	sourceAnalytics, err := s.calculateAnalyticByIndicator(ctx, exchangeName, symbol, unit, interval, calc.ByAnalytic(), indicatorDepth, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch source analytics")
	}
	return s.OscillatorByAnalytics(ctx, sourceAnalytics, name, depth)
}

func (s *Service) calculateOscillatorByIndicatorToDate(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int, datetime time.Time) ([]Analytic, error) {
	calc, has := s.calculatorsByAnalytic[name]
	if !has {
		return nil, nil
	}
	sourceAnalytics, err := s.calculateAnalyticByIndicatorToDate(ctx, exchangeName, symbol, unit, interval, calc.ByAnalytic(), indicatorDepth, depth, datetime)
	if err != nil {
		return nil, errors.Wrap(err, "fetch source analytics")
	}
	return s.OscillatorByAnalytics(ctx, sourceAnalytics, name, depth)
}

func (s *Service) calculateAnalyticByIndicator(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int) ([]Analytic, error) {
	calc, hasByIndicator := s.calculatorsByIndicator[name]
	if !hasByIndicator {
		return nil, nil
	}
	indicators, err := s.indicatorService.CalculateLastSequence(ctx, exchangeName, symbol, unit, interval, calc.ByIndicator(), indicatorDepth, 100)
	if err != nil {
		return nil, errors.Wrap(err, "fetch indicators")
	}
	return s.AnalyticByIndicators(ctx, indicators, name, depth)
}
func (s *Service) calculateAnalyticByIndicatorToDate(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int, datetime time.Time) ([]Analytic, error) {
	calc, hasByIndicator := s.calculatorsByIndicator[name]
	if !hasByIndicator {
		return nil, nil
	}
	indicators, err := s.indicatorService.LastSequenceToDate(ctx, exchangeName, symbol, unit, interval, calc.ByIndicator(), indicatorDepth, 100, datetime)
	if err != nil {
		return nil, errors.Wrap(err, "fetch indicators")
	}
	return s.AnalyticByIndicators(ctx, indicators, name, depth)
}

func (s *Service) DeleteOldRows(ctx context.Context, limit int) error {
	groups, err := s.repo.UniqGroups(ctx)
	if err != nil {
		return errors.Wrap(err, "fetch from storage")
	}
	for _, group := range groups {
		lastRaw, err := s.repo.LastInGroup(ctx, group, limit)
		if err != nil {
			return err
		}
		if lastRaw == nil {
			continue
		}
		if err := s.repo.DeleteByGroup(ctx, group, lastRaw.Datetime); err != nil {
			return err
		}
	}
	return nil
}
