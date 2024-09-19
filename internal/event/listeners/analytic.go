package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
)

type Analytic struct {
	analysisHandler *service.AnalysisHandler
}

func NewAnalytic(analysisHandler *service.AnalysisHandler) dispatcher.Listener[analysis.Analytic] {
	return &Analytic{analysisHandler: analysisHandler}
}

func (a *Analytic) Handle(e dispatcher.Event[analysis.Analytic]) {
	a.analysisHandler.HandleAnalytic(context.TODO(), e.Body)
}
