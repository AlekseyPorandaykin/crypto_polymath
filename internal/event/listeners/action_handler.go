package listeners

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
)

type ActionHandler struct {
	clientSender *grpc.ActionHandler
}

func NewActionHandler(clientSender *grpc.ActionHandler) *ActionHandler {
	return &ActionHandler{clientSender: clientSender}
}

func (h *ActionHandler) Handle(event dispatcher.Event[domain.ActionBody]) {
	h.clientSender.Accept(event.Name, event.Body)
}
