package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/pkg/errors"
)

type Service struct {
	calculators map[string]Calculator
	repo        Repository
}

func NewService(repo Repository) *Service {
	return &Service{calculators: make(map[string]Calculator), repo: repo}
}

func (s *Service) AddCalculator(calc Calculator) {
	s.calculators[calc.Name()] = calc
}

func (s *Service) CalculateByIndicator(ctx context.Context, indicator domain.Indicator) error {
	for _, calc := range s.calculators {
		if indicator.Name != calc.ByIndicator() {
			continue
		}
		if !calc.SupportInterval(indicator.Interval) {
			continue
		}
		analytics, err := calc.Calculate(ctx, indicator)
		if err != nil {
			return err
		}
		for _, analytic := range analytics {
			if err := s.repo.Save(ctx, analytic); err != nil {
				return err
			}
		}
	}
	return nil
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
