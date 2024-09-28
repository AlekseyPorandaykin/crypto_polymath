package service

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"go.uber.org/zap"
)

type AnalysisHandler struct {
	analysisService    *analysis.Service
	analyticDispatcher *dispatcher.Dispatcher[analysis.Analytic]
}

func NewAnalysisHandler(analysisService *analysis.Service, analyticDispatcher *dispatcher.Dispatcher[analysis.Analytic]) *AnalysisHandler {
	return &AnalysisHandler{analysisService: analysisService, analyticDispatcher: analyticDispatcher}
}

func (a *AnalysisHandler) HandleIndicator(ctx context.Context, data domain.Indicator) {
	analytics, err := a.analysisService.CalculateByIndicator(ctx, data)
	if err != nil {
		zap.L().Error("failed to calculate by indicator", zap.Error(err))
		return
	}
	for _, analyticItem := range analytics {
		a.analyticDispatcher.Dispatch(dispatcher.Event[analysis.Analytic]{
			Name: domain.CreatedAnalyticEvent,
			Body: analyticItem,
		})
	}
}

func (a *AnalysisHandler) HandleAnalytic(ctx context.Context, data analysis.Analytic) {
	analytics, err := a.analysisService.CalculateByAnalytic(ctx, data)
	if err != nil {
		zap.L().Error("failed to calculate by analytic", zap.Error(err))
		return
	}
	for _, analyticItem := range analytics {
		a.analyticDispatcher.Dispatch(dispatcher.Event[analysis.Analytic]{
			Name: domain.CreatedAnalyticEvent,
			Body: analyticItem,
		})
	}
}
