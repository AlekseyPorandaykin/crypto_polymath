package service

import (
	"context"
	"sync"

	v5 "github.com/AlekseyPorandaykin/crypto-exchanges/exchange/bybit/v5"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/impl/view"
	"github.com/AlekseyPorandaykin/go-kit/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/spf13/viper"
)

type DictionaryRepositories struct {
	Analysis   analysis.Repository
	Indicators indicator.Repository
	Symbols    candlestick.Repository
}

type Service struct {
	repositories DictionaryRepositories

	dictionaryCache *view.DictionaryModel
	mu              sync.Mutex
}

func NewService(repositories DictionaryRepositories) *Service {
	return &Service{repositories: repositories}
}

func (s *Service) Collect(ctx context.Context) error {
	dictionary, err := s.collectDictionary(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.dictionaryCache = &dictionary
	s.mu.Unlock()
	return nil
}

func (s *Service) Dictionary(ctx context.Context) (view.DictionaryModel, error) {
	if s.dictionaryCache != nil {
		return *s.dictionaryCache, nil
	}
	dictionary, err := s.collectDictionary(ctx)
	if err != nil {
		return view.DictionaryModel{}, err
	}
	s.mu.Lock()
	s.dictionaryCache = &dictionary
	s.mu.Unlock()
	return dictionary, nil
}

func (s *Service) collectDictionary(ctx context.Context) (view.DictionaryModel, error) {
	unitIntervals := make([]view.IntervalModel, 0, 10)
	analysisData := make([]view.AnalysisModel, 0, 10)
	indicatorData := make([]view.IndicatorModel, 0, 10)
	symbols := viper.GetStringSlice("load.symbols")
	allDepths := viper.GetIntSlice("candlestick.depths")
	allIndicatorDepths := viper.GetIntSlice("candlestick.depths")
	symbolsData, err := s.repositories.Symbols.AllSymbols(ctx)
	if err != nil {
		return view.DictionaryModel{}, err
	}
	symbols = append(symbols, symbolsData...)
	analyticInfoData, err := s.repositories.Analysis.AllAnalyticInfo(ctx)
	if err != nil {
		return view.DictionaryModel{}, err
	}
	for nameAnalyticInfo, analyticInfoItem := range analyticInfoData {
		depths := make([]int, 0, 100)
		indicatorDepths := make([]int, 0, 100)
		for _, item := range analyticInfoItem {
			depths = append(depths, item.Depth)
			indicatorDepths = append(indicatorDepths, item.IndicatorDepth)
		}
		analysisData = append(analysisData, view.AnalysisModel{
			Name:           nameAnalyticInfo,
			Description:    domain.IndicatorDescriptions[nameAnalyticInfo],
			Depth:          util.UniqSlice(depths),
			IndicatorDepth: util.UniqSlice(indicatorDepths),
		})
		allDepths = append(allDepths, util.UniqSlice(depths)...)
		allIndicatorDepths = append(allIndicatorDepths, util.UniqSlice(indicatorDepths)...)
	}
	indicatorInfoData, err := s.repositories.Indicators.AllIndicatorInfo(ctx)
	if err != nil {
		return view.DictionaryModel{}, err
	}
	for nameIndicatorInfo, indicatorInfoItem := range indicatorInfoData {
		depths := make([]int, 0, 100)
		for _, item := range indicatorInfoItem {
			depths = append(depths, item.Depth)
		}
		indicatorData = append(indicatorData, view.IndicatorModel{
			Name:        nameIndicatorInfo,
			Description: domain.IndicatorDescriptions[nameIndicatorInfo],
			Depth:       util.UniqSlice(depths),
		})
	}
	unitIntervals = append(unitIntervals, view.IntervalModel{
		Unit:   string(domain.HourUnit),
		Values: viper.GetIntSlice("candlestick.hours"),
	})
	unitIntervals = append(unitIntervals, view.IntervalModel{
		Unit:   string(domain.MinuteUnit),
		Values: viper.GetIntSlice("candlestick.minutes"),
	})
	unitIntervals = append(unitIntervals, view.IntervalModel{
		Unit:   string(domain.DayUnit),
		Values: []int{1},
	})
	unitIntervals = append(unitIntervals, view.IntervalModel{
		Unit:   string(domain.WeekUnit),
		Values: []int{1},
	})
	unitIntervals = append(unitIntervals, view.IntervalModel{
		Unit:   string(domain.MonthUnit),
		Values: []int{1},
	})
	allDepths = util.UniqSlice(allDepths)
	slice.Sort(allDepths)
	allIndicatorDepths = util.UniqSlice(allIndicatorDepths)
	slice.Sort(allIndicatorDepths)
	return view.DictionaryModel{
		Analysis:       analysisData,
		Depths:         util.UniqSlice(allDepths),
		Exchanges:      []string{v5.ExchangeName},
		IndicatorDepth: util.UniqSlice(allIndicatorDepths),
		Indicators:     indicatorData,
		Intervals:      unitIntervals,
		Symbols:        util.UniqSlice(symbols),
		Units: []string{
			string(domain.MinuteUnit),
			string(domain.HourUnit),
			string(domain.DayUnit),
			string(domain.WeekUnit),
			string(domain.MonthUnit),
		},
	}, nil
}
