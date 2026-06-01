package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kanengo/ku/mqx"
	"github.com/kanengo/ku/snowflakex"
	kimgatev1 "github.com/project-kgo/kim-gate/proto/kimgate/v1"
	"github.com/project-kgo/kim/internal/event"
	"github.com/project-kgo/kim/internal/model"
	kimv1 "github.com/project-kgo/kim/proto/kim/v1"
	"google.golang.org/protobuf/proto"
)

func TestC2CConversationID(t *testing.T) {
	tests := []struct {
		name       string
		senderID   int64
		receiverID int64
		want       string
		wantErr    bool
	}{
		{name: "orders ids ascending", senderID: 10, receiverID: 2, want: "2-10"},
		{name: "keeps ascending ids", senderID: 2, receiverID: 10, want: "2-10"},
		{name: "rejects empty participant", senderID: 0, receiverID: 10, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := C2CConversationID(tt.senderID, tt.receiverID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("C2CConversationID returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("conversation id = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandleC2CMessagePersistsThenPushesReceiver(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)
	evt := event.MessageEvent{
		MessageID:      1001,
		ConversationID: "client-wrong",
		SenderID:       "10",
		ReceiverID:     "2",
		Content:        "hello",
		Type:           "text",
		CreatedAt:      createdAt.UnixMilli(),
	}
	raw, err := sonic.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	store := &fakeC2CStore{}
	gateway := &fakeGatewayPusher{}
	pusher := NewMessagePusher(gateway)
	node, _ := snowflakex.NewNode(1, 0)
	consumer := NewConsumer(slog.Default(), store, &fakeConsumerPubSub{}, pusher, node)

	if err := consumer.handleC2CMessage(ctx, &mqx.Message{Data: raw}); err != nil {
		t.Fatalf("handleC2CMessage returned error: %v", err)
	}

	if len(store.records) != 1 {
		t.Fatalf("persist calls = %d, want 1", len(store.records))
	}
	got := store.records[0]
	if got.Message.ConversationID != "2-10" {
		t.Fatalf("message conversation_id = %q, want 2-10", got.Message.ConversationID)
	}
	if got.Message.Content != `"hello"` {
		t.Fatalf("stored content = %q, want JSON string", got.Message.Content)
	}
	if len(got.SyncMails) != 2 {
		t.Fatalf("sync mail count = %d, want 2", len(got.SyncMails))
	}
	if got.SyncMails[0].UserID != 10 || got.SyncMails[1].UserID != 2 {
		t.Fatalf("sync mail users = %d,%d want sender then receiver", got.SyncMails[0].UserID, got.SyncMails[1].UserID)
	}
	if len(got.Conversations) != 2 {
		t.Fatalf("conversation count = %d, want 2", len(got.Conversations))
	}
	if got.Conversations[0].UserID != 10 || got.Conversations[0].Unread != 0 {
		t.Fatalf("sender conversation = %+v", got.Conversations[0])
	}
	if got.Conversations[1].UserID != 2 || got.Conversations[1].Unread != 1 {
		t.Fatalf("receiver conversation = %+v", got.Conversations[1])
	}
	if len(gateway.requests) != 1 {
		t.Fatalf("gateway calls = %d, want 1", len(gateway.requests))
	}
	req := gateway.requests[0]
	if len(req.UserIds) != 1 || req.UserIds[0] != "2" {
		t.Fatalf("gateway user ids = %v, want [2]", req.UserIds)
	}
	if req.Method != C2CMessagePushMethod {
		t.Fatalf("gateway method = %q, want %q", req.Method, C2CMessagePushMethod)
	}
	var pushed kimv1.C2CMessagePush
	if err := proto.Unmarshal(req.Payload, &pushed); err != nil {
		t.Fatalf("unmarshal pushed payload: %v", err)
	}
	if pushed.ConversationId != "2-10" || pushed.MessageId != evt.MessageID {
		t.Fatalf("pushed payload = %+v", pushed)
	}
	if pushed.SenderId != "10" || pushed.ReceiverId != "2" || pushed.Content != "hello" || pushed.Type != "text" {
		t.Fatalf("pushed payload content = %+v", pushed)
	}
}

func TestHandleC2CMessageReturnsGatewayErrorAfterPersist(t *testing.T) {
	ctx := context.Background()
	evt := event.MessageEvent{
		MessageID:  1002,
		SenderID:   "1",
		ReceiverID: "2",
		Content:    "hello",
		Type:       "text",
		CreatedAt:  time.Now().UnixMilli(),
	}
	raw, _ := sonic.Marshal(evt)
	store := &fakeC2CStore{}
	gateway := &fakeGatewayPusher{err: errors.New("gateway unavailable")}
	pusher := NewMessagePusher(gateway)
	node, _ := snowflakex.NewNode(1, 0)
	consumer := NewConsumer(slog.Default(), store, &fakeConsumerPubSub{}, pusher, node)

	err := consumer.handleC2CMessage(ctx, &mqx.Message{Data: raw})
	if err == nil {
		t.Fatal("expected gateway error")
	}
	if len(store.records) != 1 {
		t.Fatalf("persist calls = %d, want 1", len(store.records))
	}
}

type fakeC2CStore struct {
	records []C2CMessageRecord
	err     error
}

func (f *fakeC2CStore) SaveC2CMessage(ctx context.Context, record C2CMessageRecord) error {
	if f.err != nil {
		return f.err
	}
	f.records = append(f.records, record)
	return nil
}

type fakeGatewayPusher struct {
	requests []*kimgatev1.SendToUsersRequest
	err      error
}

func (f *fakeGatewayPusher) SendToUsers(ctx context.Context, req *kimgatev1.SendToUsersRequest) (*kimgatev1.SendResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.requests = append(f.requests, req)
	return &kimgatev1.SendResponse{}, nil
}

type fakeConsumerPubSub struct{}

func (f *fakeConsumerPubSub) Publish(context.Context, *mqx.PublishRequest) error {
	return nil
}

func (f *fakeConsumerPubSub) Subscribe(context.Context, *mqx.SubscribeRequest) (mqx.Subscription, error) {
	return nil, nil
}

func (f *fakeConsumerPubSub) Close() error {
	return nil
}

var _ c2cMessageStore = (*fakeC2CStore)(nil)
var _ gatewayPusher = (*fakeGatewayPusher)(nil)
var _ c2cMessagePusher = (*MessagePusher)(nil)
var _ = model.UserSyncMail{}
