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

type AnalyticInfoRepository interface {
	AllAnalyticInfo(ctx context.Context) (map[string][]AnalyticInfoModel, error)
}

type IndicatorInfoRepository interface {
	AllIndicatorInfoModel(ctx context.Context) (map[string][]IndicatorInfoModel, error)
}
