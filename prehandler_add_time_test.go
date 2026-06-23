package serverlog

import (
	"testing"
	"time"
)

func TestAddTimePreHandler(t *testing.T) {
	data := M{"key": "val"}
	result := AddTimePreHandler(InfoLevel, "msg", data)

	if result == nil {
		t.Fatal("AddTimePreHandler returned nil")
	}
	if _, ok := result["Time"]; !ok {
		t.Error("expected 'Time' key in result")
	}
	if ts, ok := result["Time"].(int64); ok {
		now := time.Now().Unix()
		if ts < now-1 || ts > now+1 {
			t.Errorf("Time = %d, want around %d", ts, now)
		}
	} else {
		t.Errorf("Time is not int64, got %T", result["Time"])
	}
	// Original key should be preserved
	if result["key"] != "val" {
		t.Errorf("key = %v, want 'val'", result["key"])
	}
}

func TestAddTimePreHandler_NilData(t *testing.T) {
	result := AddTimePreHandler(InfoLevel, "msg", nil)
	if result != nil {
		t.Errorf("expected nil for nil data, got %v", result)
	}
}
