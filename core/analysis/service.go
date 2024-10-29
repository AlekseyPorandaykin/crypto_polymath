package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/pkg/errors"
	"time"
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
	calculatedAnalytics := make([]Analytic, 0, len(s.calculatorsByIndicator))
	for analyticName := range s.calculatorsByIndicator {
		for _, depth := range s.depths {
			analytic, err := s.analyticByIndicator(ctx, indicator, analyticName, depth)
			if err != nil {
				return nil, err
			}
			if analytic == nil {
				continue
			}
			calculatedAnalytics = append(calculatedAnalytics, *analytic)
		}
	}
	return calculatedAnalytics, nil
}

func (s *Service) AnalyticByIndicators(ctx context.Context, indicators []domain.Indicator, analyticName string, depth int) ([]Analytic, error) {
	result := make([]Analytic, 0, len(indicators))
	for _, indicatorItem := range indicators {
		data, err := s.analyticByIndicator(ctx, indicatorItem, analyticName, depth)
		if err != nil {
			return nil, err
		}
		if data == nil {
			continue
		}
		result = append(result, *data)
	}
	return result, nil
}

func (s *Service) analyticByIndicator(ctx context.Context, indicator domain.Indicator, analyticName string, depth int) (*Analytic, error) {
	calc, has := s.calculatorsByIndicator[analyticName]
	if !has {
		return nil, nil
	}
	storageData, errStorage := s.repo.Find(
		ctx,
		analyticName,
		indicator.Exchange,
		indicator.Symbol,
		indicator.Unit,
		indicator.Datetime,
		indicator.Interval,
		indicator.Depth,
		depth,
	)
	if errStorage != nil {
		return nil, errors.Wrap(errStorage, "fetch from storage")
	}
	if storageData != nil {
		return storageData, nil
	}
	if indicator.Name != calc.ByIndicator() {
		return nil, nil
	}
	if !calc.SupportInterval(indicator.Interval) {
		return nil, nil
	}
	if !calc.SupportDepth(depth) {
		return nil, nil
	}
	analytic, err := calc.Calculate(ctx, indicator, depth)
	if err != nil {
		return nil, err
	}
	if analytic == nil {
		return nil, nil
	}
	if errSave := s.repo.Save(ctx, *analytic); errSave != nil {
		return nil, errors.Wrap(errSave, "save analytic")
	}
	return analytic, nil
}

func (s *Service) CalculateByAnalytic(ctx context.Context, data Analytic) ([]Analytic, error) {
	calculatedAnalytics := make([]Analytic, 0, len(s.calculatorsByIndicator))
	for _, calc := range s.calculatorsByAnalytic {
		for _, depth := range s.depths {
			oscillator, err := s.oscillatorByAnalytic(ctx, data, calc.Name(), depth)
			if err != nil {
				return nil, err
			}
			if oscillator == nil {
				continue
			}
			calculatedAnalytics = append(calculatedAnalytics, *oscillator)
		}
	}
	return calculatedAnalytics, nil
}

func (s *Service) OscillatorByAnalytics(ctx context.Context, data []Analytic, oscillatorName string, depth int) ([]Analytic, error) {
	result := make([]Analytic, 0, len(data))
	for _, analyticItem := range data {
		oscillator, err := s.oscillatorByAnalytic(ctx, analyticItem, oscillatorName, depth)
		if err != nil {
			return nil, err
		}
		if oscillator == nil {
			continue
		}
		result = append(result, *oscillator)
	}
	return result, nil
}

// analyticByAnalytic - рассчитываем сложную аналитику на основе другой аналитики
func (s *Service) oscillatorByAnalytic(ctx context.Context, data Analytic, oscillatorName string, depth int) (*Analytic, error) {
	calc, has := s.calculatorsByAnalytic[oscillatorName]
	if !has {
		return nil, nil
	}
	storageData, err := s.repo.Find(ctx, oscillatorName, data.Exchange, data.Symbol, data.Unit, data.Datetime, data.Interval, data.Depth, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	if storageData != nil {
		return storageData, nil
	}
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
	analyticData, err := calc.Calculate(ctx, data, depth)
	if err != nil {
		return nil, errors.Wrap(err, "calculate oscillatorByAnalytic")
	}
	if analyticData == nil {
		return nil, nil
	}
	if errSave := s.repo.Save(ctx, *analyticData); errSave != nil {
		return nil, errors.Wrap(errSave, "save oscillatorByAnalytic")
	}
	return analyticData, nil
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
