package cmdutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUnmarshalInputKeepsLargeIntegers(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.json")
	if err := os.WriteFile(file, []byte(`{"user_id": 9007199254740993, "score": 1.5}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var body interface{}
	if err := UnmarshalInput(file, &body); err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"score":1.5,"user_id":9007199254740993}` {
		t.Fatalf("re-encoded body = %s", got)
	}
}

func TestUnmarshalInputRejectsTrailingData(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.json")
	if err := os.WriteFile(file, []byte(`{"a":1} {"b":2}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var body interface{}
	if err := UnmarshalInput(file, &body); err == nil {
		t.Fatalf("expected an error, got %v", body)
	}
}
