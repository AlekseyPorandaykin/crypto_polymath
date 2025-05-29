package grpc

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/grpc/action"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
	"time"
)

type ActionEvent struct {
	Name string
	Body domain.LoadedCandlesticksActionBody
}
type ActionHandler struct {
	action.UnimplementedActionServiceServer
	mu           sync.Mutex
	actionEvents map[uuid.UUID][]ActionEvent

	muActiveSubscribers sync.Mutex
	activeSubscribers   int32
}

func NewActionHandler() *ActionHandler {
	return &ActionHandler{actionEvents: make(map[uuid.UUID][]ActionEvent)}
}

func (h *ActionHandler) Accept(name string, action domain.LoadedCandlesticksActionBody) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.hasSubscribers() {
		if len(h.actionEvents) > 0 {
			h.actionEvents = make(map[uuid.UUID][]ActionEvent)
		}
		return
	}
	for uniqSubscriberKey := range h.actionEvents {
		h.actionEvents[uniqSubscriberKey] = append(h.actionEvents[uniqSubscriberKey], ActionEvent{Name: name, Body: action})
	}
}

func (h *ActionHandler) receive(uniqKey uuid.UUID) []ActionEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	events, ok := h.actionEvents[uniqKey]
	if !ok {
		return nil
	}
	h.actionEvents[uniqKey] = make([]ActionEvent, 0, 100)
	return events
}

func (h *ActionHandler) flushEvents() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.actionEvents = make(map[uuid.UUID][]ActionEvent)
}

func (h *ActionHandler) hasSubscribers() bool {
	h.muActiveSubscribers.Lock()
	defer h.muActiveSubscribers.Unlock()
	return h.activeSubscribers > 0
}

func (h *ActionHandler) addSubscriber(uniqKey uuid.UUID) {
	h.muActiveSubscribers.Lock()
	h.activeSubscribers++
	h.muActiveSubscribers.Unlock()

	h.mu.Lock()
	h.actionEvents[uniqKey] = make([]ActionEvent, 0, 100)
	defer h.mu.Unlock()
}
func (h *ActionHandler) removeSubscriber(uniqKey uuid.UUID) {
	h.muActiveSubscribers.Lock()
	defer h.muActiveSubscribers.Unlock()
	if h.activeSubscribers > 0 {
		delete(h.actionEvents, uniqKey)
		h.activeSubscribers--
	}
	if h.activeSubscribers == 0 {
		h.flushEvents()
	}
}

func (h *ActionHandler) StreamActions(req *action.ActionRequest, resp grpc.ServerStreamingServer[action.Action]) error {
	uniqKey := uuid.New()
	h.addSubscriber(uniqKey)
	defer h.removeSubscriber(uniqKey)
	ticker := time.NewTicker(1)
	defer ticker.Stop()
	for {
		select {
		case <-resp.Context().Done():
			return resp.Context().Err()
		case <-ticker.C:
			ticker.Stop()
			for _, item := range h.receive(uniqKey) {
				if req.GetAction() != "" && req.GetAction() != item.Name {
					continue
				}
				if errItem := resp.Send(&action.Action{
					Action:   item.Name,
					Exchange: item.Body.Exchange,
					Symbol:   item.Body.Symbol,
					Unit:     string(item.Body.Unit),
					Interval: int32(item.Body.Interval),
					CreatedAt: &timestamppb.Timestamp{
						Seconds: int64(item.Body.CreatedAt.UnixMilli() / 1000),
						Nanos:   int32(item.Body.CreatedAt.UnixNano()),
					},
					DurationInMs: float32(item.Body.Duration.Milliseconds()),
				}); errItem != nil {
					return errItem
				}
			}
			ticker.Reset(time.Second * 5)
		}
	}
}
