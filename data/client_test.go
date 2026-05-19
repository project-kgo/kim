package data

import "testing"

const (
	validRedisDSN = "redis://localhost:6379/0"
	validDBDSN    = "postgres://kim:secret@localhost:5432/kim?sslmode=disable"
)

func TestNewClientRequiresRedisDSN(t *testing.T) {
	if _, err := NewClient(Config{DBDSN: validDBDSN}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientRequiresDBDSN(t *testing.T) {
	if _, err := NewClient(Config{RedisDSN: validRedisDSN}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientRejectsInvalidRedisDSN(t *testing.T) {
	if _, err := NewClient(Config{RedisDSN: "://bad", DBDSN: validDBDSN}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientCreatesClients(t *testing.T) {
	client, err := NewClient(Config{
		RedisDSN: " " + validRedisDSN + " ",
		DBDSN:    " " + validDBDSN + " ",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	if client.Redis == nil {
		t.Fatal("Redis client is nil")
	}
	if client.DB == nil {
		t.Fatal("DB client is nil")
	}
	if client.DB.DriverName() != "pgx" {
		t.Fatalf("DB driver = %q, want %q", client.DB.DriverName(), "pgx")
	}
}
