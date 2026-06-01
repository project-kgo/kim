package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kanengo/ku/mqx"
	"github.com/kanengo/ku/snowflakex"
	"github.com/project-kgo/kim/internal/event"
	"github.com/project-kgo/kim/internal/model"
)

func TestSendGeneratesC2CConversationID(t *testing.T) {
	pubsub := &recordingPubSub{}
	node, _ := snowflakex.NewNode(1, 0)
	svc := NewMessageService(slog.Default(), node, pubsub)

	resp, err := svc.Send(context.Background(), model.SendMessageRequest{
		ConversationID: "client-wrong",
		SenderID:       "10",
		ReceiverID:     "2",
		Content:        "hello",
		MessageType:    "text",
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if resp.MessageID == "" || resp.CreatedAt == "" {
		t.Fatalf("empty response: %+v", resp)
	}
	if len(pubsub.published) != 1 {
		t.Fatalf("published count = %d, want 1", len(pubsub.published))
	}
	var evt event.MessageEvent
	if err := sonic.Unmarshal(pubsub.published[0].Data, &evt); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if evt.ConversationID != "2-10" {
		t.Fatalf("conversation_id = %q, want 2-10", evt.ConversationID)
	}
}

type recordingPubSub struct {
	published []*mqx.PublishRequest
}

func (r *recordingPubSub) Publish(_ context.Context, req *mqx.PublishRequest) error {
	r.published = append(r.published, req)
	return nil
}

func (r *recordingPubSub) Subscribe(_ context.Context, _ *mqx.SubscribeRequest) (mqx.Subscription, error) {
	return nil, nil
}

func (r *recordingPubSub) Close() error {
	return nil
}
