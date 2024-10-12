package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/pkg/errors"
)

type Service struct {
	calculatorsByIndicator map[string]CalculatorByIndicator
	calculatorsByAnalytic  map[string]CalculatorByAnalytic
	repo                   Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		calculatorsByIndicator: make(map[string]CalculatorByIndicator),
		calculatorsByAnalytic:  make(map[string]CalculatorByAnalytic),
		repo:                   repo,
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
	for _, calc := range s.calculatorsByIndicator {
		if indicator.Name != calc.ByIndicator() {
			continue
		}
		if !calc.SupportInterval(indicator.Interval) {
			continue
		}
		analytics, err := calc.Calculate(ctx, indicator)
		if err != nil {
			return nil, err
		}
		for _, analytic := range analytics {
			if err := s.repo.Save(ctx, analytic); err != nil {
				return nil, err
			}
		}
		calculatedAnalytics = append(calculatedAnalytics, analytics...)
	}
	return calculatedAnalytics, nil
}

func (s *Service) CalculateByAnalytic(ctx context.Context, data Analytic) ([]Analytic, error) {
	calculatedAnalytics := make([]Analytic, 0, len(s.calculatorsByIndicator))
	for _, calc := range s.calculatorsByAnalytic {
		if data.Name != calc.ByAnalytic() {
			continue
		}
		if !calc.SupportInterval(data.Interval) {
			continue
		}
		analytics, err := calc.Calculate(ctx, data)
		if err != nil {
			return nil, err
		}
		for _, analytic := range analytics {
			if err := s.repo.Save(ctx, analytic); err != nil {
				return nil, err
			}
		}
		calculatedAnalytics = append(calculatedAnalytics, analytics...)
	}
	return calculatedAnalytics, nil
}

func (s *Service) Analytics(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int,
) ([]Analytic, error) {
	data, err := s.repo.Last(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	return data, nil
}

func (s *Service) SequenceAnalytics(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int,
) ([]Analytic, error) {
	data, err := s.Analytics(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	if !isCorrectSequence(data) {
		return nil, nil
	}
	return data, nil
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
