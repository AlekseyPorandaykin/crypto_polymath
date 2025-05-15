package application

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	adapter_exchange "github.com/AlekseyPorandaykin/crypto_polymath/internal/adapters/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/view"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/spf13/viper"
	"sync"
)

type Service struct {
	analyticRepository  view.AnalyticInfoRepository
	indicatorRepository view.IndicatorInfoRepository
	symbolRepository    view.SymbolRepository

	dictionaryCache *view.DictionaryModel
	mu              sync.Mutex
}

func NewService(
	analyticRepository view.AnalyticInfoRepository,
	indicatorRepository view.IndicatorInfoRepository,
	symbolRepository view.SymbolRepository,
) *Service {
	return &Service{
		analyticRepository:  analyticRepository,
		indicatorRepository: indicatorRepository,
		symbolRepository:    symbolRepository,
	}
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
	symbolsData, err := s.symbolRepository.AllSymbols(ctx)
	if err != nil {
		return view.DictionaryModel{}, err
	}
	symbols = append(symbols, symbolsData...)
	analyticInfoData, err := s.analyticRepository.AllAnalyticInfo(ctx)
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
	indicatorInfoData, err := s.indicatorRepository.AllIndicatorInfoModel(ctx)
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
		Exchanges:      []string{adapter_exchange.BybitExchange},
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
