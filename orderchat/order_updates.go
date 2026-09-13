package orderchat

import (
	"context"
	"encoding/json"
	"fmt"
)

type Stage string

const (
	StageCheckoutPending Stage = "checkout.pending"
	StageCheckoutPaid    Stage = "checkout.paid"
	StagePacking         Stage = "fulfillment.packing"
	StageShipped         Stage = "fulfillment.shipped"
	StageReceiptIssued   Stage = "receipt.issued"
)

var nextStage = map[Stage]Stage{
	StageCheckoutPending: StageCheckoutPaid,
	StageCheckoutPaid:    StagePacking,
	StagePacking:         StageShipped,
	StageShipped:         StageReceiptIssued,
}

type OrderUpdate struct {
	OrderID   string `json:"order_id"`
	AccountID string `json:"account_id"`
	From      Stage  `json:"from"`
	To        Stage  `json:"to"`
	Message   string `json:"message"`
	EventID   string `json:"event_id"`
}

func (u OrderUpdate) Validate() error {
	if u.OrderID == "" || u.AccountID == "" || u.EventID == "" {
		return fmt.Errorf("order_id, account_id, and event_id are required")
	}
	if nextStage[u.From] != u.To {
		return fmt.Errorf("invalid order transition %q to %q", u.From, u.To)
	}
	return nil
}

type Realtime interface {
	CreateChannel(context.Context, string, string) error
	IssueToken(context.Context, string, string) (json.RawMessage, error)
	Publish(context.Context, string, string, any, string, string) error
	Presence(context.Context, string) (json.RawMessage, error)
}

type OrderRoom struct{ realtime Realtime }

func NewOrderRoom(realtime Realtime) *OrderRoom { return &OrderRoom{realtime: realtime} }

func Channel(orderID string) string { return "order:" + orderID }

func (r *OrderRoom) Open(ctx context.Context, orderID, clientID, requestID string) (json.RawMessage, error) {
	channel := Channel(orderID)
	if err := r.realtime.CreateChannel(ctx, channel, requestID); err != nil {
		return nil, err
	}
	return r.realtime.IssueToken(ctx, clientID, channel)
}

func (r *OrderRoom) Apply(ctx context.Context, update OrderUpdate) error {
	if err := update.Validate(); err != nil {
		return err
	}
	return r.realtime.Publish(ctx, Channel(update.OrderID), "order.updated", update, update.AccountID, update.EventID)
}

func (r *OrderRoom) Viewers(ctx context.Context, orderID string) (json.RawMessage, error) {
	return r.realtime.Presence(ctx, Channel(orderID))
}
