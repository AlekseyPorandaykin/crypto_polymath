package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"time"
)

type Analytic struct {
	analysisHandler *application.AnalysisHandler
}

func NewAnalytic(analysisHandler *application.AnalysisHandler) dispatcher.Listener[analysis.Analytic] {
	return &Analytic{analysisHandler: analysisHandler}
}

func (a *Analytic) Handle(e dispatcher.Event[analysis.Analytic]) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	a.analysisHandler.HandleAnalytic(ctx, e.Body)
}
