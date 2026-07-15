package view

import "context"

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

type DictionaryRepository interface {
	Dictionary(ctx context.Context) (DictionaryModel, error)
}
