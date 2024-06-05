package service

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

type AnalysisHandler struct {
	analysisService *analysis.Service
}

func NewAnalysisHandler(analysisService *analysis.Service) *AnalysisHandler {
	return &AnalysisHandler{analysisService: analysisService}
}

func (a *AnalysisHandler) HandleIndicator(ctx context.Context, data domain.Indicator) {
	_ = a.analysisService.CalculateByIndicator(ctx, data)
}
