package data

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestMessageDedupStoreReserveGetDeleteC2CMessage(t *testing.T) {
	ctx := context.Background()
	client := newFakeMessageDedupRedis()
	store := newMessageDedupStore(client)
	value := MessageDedupValue{
		MessageID:   "1001",
		CreatedAt:   "2026-06-03T10:00:00Z",
		Fingerprint: "fp",
	}

	ok, err := store.ReserveC2CMessage(ctx, "10", "client-1", value, 5*time.Minute)
	if err != nil {
		t.Fatalf("ReserveC2CMessage returned error: %v", err)
	}
	if !ok {
		t.Fatal("first reserve = false, want true")
	}
	if client.ttl != 5*time.Minute {
		t.Fatalf("ttl = %s, want 5m", client.ttl)
	}

	ok, err = store.ReserveC2CMessage(ctx, "10", "client-1", value, 5*time.Minute)
	if err != nil {
		t.Fatalf("duplicate ReserveC2CMessage returned error: %v", err)
	}
	if ok {
		t.Fatal("duplicate reserve = true, want false")
	}

	got, err := store.GetC2CMessage(ctx, "10", "client-1")
	if err != nil {
		t.Fatalf("GetC2CMessage returned error: %v", err)
	}
	if got != value {
		t.Fatalf("dedup value = %+v, want %+v", got, value)
	}

	if err := store.DeleteC2CMessage(ctx, "10", "client-1"); err != nil {
		t.Fatalf("DeleteC2CMessage returned error: %v", err)
	}
	ok, err = store.ReserveC2CMessage(ctx, "10", "client-1", value, 5*time.Minute)
	if err != nil {
		t.Fatalf("ReserveC2CMessage after delete returned error: %v", err)
	}
	if !ok {
		t.Fatal("reserve after delete = false, want true")
	}
}

type fakeMessageDedupRedis struct {
	values map[string]string
	ttl    time.Duration
}

func newFakeMessageDedupRedis() *fakeMessageDedupRedis {
	return &fakeMessageDedupRedis{values: make(map[string]string)}
}

func (f *fakeMessageDedupRedis) SetNX(_ context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	f.ttl = expiration
	if _, ok := f.values[key]; ok {
		return redis.NewBoolResult(false, nil)
	}
	f.values[key] = value.(string)
	return redis.NewBoolResult(true, nil)
}

func (f *fakeMessageDedupRedis) Get(_ context.Context, key string) *redis.StringCmd {
	value, ok := f.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}

func (f *fakeMessageDedupRedis) Del(_ context.Context, keys ...string) *redis.IntCmd {
	var count int64
	for _, key := range keys {
		if _, ok := f.values[key]; ok {
			delete(f.values, key)
			count++
		}
	}
	return redis.NewIntResult(count, nil)
}
