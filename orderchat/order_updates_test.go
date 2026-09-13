package orderchat

import (
	"context"
	"encoding/json"
	"testing"
)

type recordingRealtime struct {
	published bool
	channel   string
	eventID   string
}

func (r *recordingRealtime) CreateChannel(context.Context, string, string) error { return nil }
func (r *recordingRealtime) IssueToken(context.Context, string, string) (json.RawMessage, error) {
	return json.RawMessage(`{"token":"client-token"}`), nil
}
func (r *recordingRealtime) Publish(_ context.Context, channel, _ string, _ any, _ string, requestID string) error {
	r.published = true
	r.channel = channel
	r.eventID = requestID
	return nil
}
func (r *recordingRealtime) Presence(context.Context, string) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

func TestOrderRoomApply(t *testing.T) {
	tests := []struct {
		name    string
		from    Stage
		to      Stage
		wantErr bool
	}{
		{name: "paid checkout starts packing", from: StageCheckoutPaid, to: StagePacking},
		{name: "packing can ship", from: StagePacking, to: StageShipped},
		{name: "unpaid checkout cannot ship", from: StageCheckoutPending, to: StageShipped, wantErr: true},
		{name: "receipt cannot precede shipment", from: StagePacking, to: StageReceiptIssued, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			realtime := &recordingRealtime{}
			room := NewOrderRoom(realtime)
			update := OrderUpdate{OrderID: "ord_2048", AccountID: "acct_7", From: tt.from, To: tt.to, EventID: "evt_91"}
			err := room.Apply(context.Background(), update)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
			if realtime.published == tt.wantErr {
				t.Fatalf("published = %v, want %v", realtime.published, !tt.wantErr)
			}
			if !tt.wantErr && (realtime.channel != "order:ord_2048" || realtime.eventID != "evt_91") {
				t.Fatalf("publish boundary = (%q, %q)", realtime.channel, realtime.eventID)
			}
		})
	}
}
