package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"time"
)

type Indicator struct {
	analysisHandler    *application.AnalysisHandler
	analyticDispatcher dispatcher.Dispatcher[analysis.Analytic]
}

func NewIndicator(analysisHandler *application.AnalysisHandler) dispatcher.Listener[domain.Indicator] {
	return &Indicator{analysisHandler: analysisHandler}
}

func (i *Indicator) Handle(e dispatcher.Event[domain.Indicator]) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	i.analysisHandler.HandleIndicator(ctx, e.Body)
}
