package view

import "context"

type AnalyticInfoModel struct {
	Unit           string
	Interval       int
	Name           string
	Depth          int
	IndicatorDepth int
}

type IndicatorInfoModel struct {
	Unit     string
	Interval int
	Name     string
	Depth    int
}

type AnalysisModel struct {
	Depth          []int
	Name           string
	Description    string
	IndicatorDepth []int
}

type IndicatorModel struct {
	Depth       []int
	Name        string
	Description string
}

type IntervalModel struct {
	Unit   string
	Values []int
}

type DictionaryModel struct {
	Analysis       []AnalysisModel
	Depths         []int
	Exchanges      []string
	IndicatorDepth []int
	Indicators     []IndicatorModel
	Intervals      []IntervalModel
	Symbols        []string
	Units          []string
}

type AnalyticInfoRepository interface {
	AllAnalyticInfo(ctx context.Context) (map[string][]AnalyticInfoModel, error)
}

type IndicatorInfoRepository interface {
	AllIndicatorInfoModel(ctx context.Context) (map[string][]IndicatorInfoModel, error)
}

type DictionaryRepository interface {
	Dictionary(ctx context.Context) (DictionaryModel, error)
}

type SymbolRepository interface {
	AllSymbols(ctx context.Context) ([]string, error)
}
