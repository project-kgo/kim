package data

import (
	"testing"
)

const (
	validRedisDSN = "redis://localhost:6379/0"
	validDBDSN    = "postgres://kim:secret@localhost:5432/kim?sslmode=disable"
)

func TestNewRequiresRedisDSN(t *testing.T) {
	if _, err := New("", "", validDBDSN, nil); err == nil {
		t.Fatal("expected error for empty redis dsn")
	}
}

func TestNewRequiresDBDSN(t *testing.T) {
	if _, err := New(validRedisDSN, "", "", nil); err == nil {
		t.Fatal("expected error for empty db dsn")
	}
}

func TestNewRejectsInvalidRedisDSN(t *testing.T) {
	if _, err := New("://bad", "", validDBDSN, nil); err == nil {
		t.Fatal("expected error for invalid redis dsn")
	}
}

func TestCloseNilSafe(t *testing.T) {
	var d *Data
	if err := d.Close(); err != nil {
		t.Fatalf("Close on nil returned error: %v", err)
	}
}
