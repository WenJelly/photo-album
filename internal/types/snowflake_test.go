package types

import (
	"encoding/json"
	"testing"
)

func TestSnowflakeIDUnmarshalJSONSupportsStringAndNumber(t *testing.T) {
	t.Parallel()

	const rawID = "1921565896585154562"

	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "string payload",
			payload: `{"id":"1921565896585154562"}`,
		},
		{
			name:    "number payload",
			payload: `{"id":1921565896585154562}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req struct {
				ID SnowflakeID `json:"id"`
			}
			if err := json.Unmarshal([]byte(tt.payload), &req); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if got := req.ID.String(); got != rawID {
				t.Fatalf("json.Unmarshal() id = %q, want %q", got, rawID)
			}
		})
	}
}

func TestSnowflakeIDMarshalJSONUsesString(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(struct {
		ID SnowflakeID `json:"id"`
	}{
		ID: NewSnowflakeID(1921565896585154562),
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	const want = `{"id":"1921565896585154562"}`
	if string(body) != want {
		t.Fatalf("json.Marshal() = %s, want %s", body, want)
	}
}

func TestSnowflakeIDUnmarshalText(t *testing.T) {
	t.Parallel()

	var id SnowflakeID
	if err := id.UnmarshalText([]byte("1921565896585154562")); err != nil {
		t.Fatalf("UnmarshalText() error = %v", err)
	}
	if got := id.String(); got != "1921565896585154562" {
		t.Fatalf("UnmarshalText() = %q, want %q", got, "1921565896585154562")
	}
}
