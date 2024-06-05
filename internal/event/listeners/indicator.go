package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
)

type Indicator struct {
	analysisHandler *service.AnalysisHandler
}

func NewIndicator(analysisHandler *service.AnalysisHandler) dispatcher.Listener[domain.Indicator] {
	return &Indicator{analysisHandler: analysisHandler}
}

func (i *Indicator) Handle(e dispatcher.Event[domain.Indicator]) {
	i.analysisHandler.HandleIndicator(context.TODO(), e.Body)
}
