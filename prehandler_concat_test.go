package serverlog

import (
	"testing"
)

func TestConcatPreHandler(t *testing.T) {
	callOrder := []string{}

	h1 := func(level Level, msg string, data M) M {
		callOrder = append(callOrder, "h1")
		if data == nil {
			data = M{}
		}
		data["h1"] = true
		return data
	}
	h2 := func(level Level, msg string, data M) M {
		callOrder = append(callOrder, "h2")
		data["h2"] = true
		return data
	}
	h3 := func(level Level, msg string, data M) M {
		callOrder = append(callOrder, "h3")
		data["h3"] = true
		return data
	}

	concat := ConcatPreHandler(h1, h2, h3)
	result := concat(InfoLevel, "msg", M{})

	// Should call in order
	if len(callOrder) != 3 {
		t.Fatalf("expected 3 handlers called, got %d", len(callOrder))
	}
	if callOrder[0] != "h1" || callOrder[1] != "h2" || callOrder[2] != "h3" {
		t.Errorf("wrong call order: %v", callOrder)
	}

	// All handlers should have modified data
	if result["h1"] != true {
		t.Error("h1 not applied")
	}
	if result["h2"] != true {
		t.Error("h2 not applied")
	}
	if result["h3"] != true {
		t.Error("h3 not applied")
	}
}

func TestConcatPreHandler_Empty(t *testing.T) {
	concat := ConcatPreHandler()
	result := concat(InfoLevel, "msg", M{"key": "val"})
	if result["key"] != "val" {
		t.Error("empty concat should return data unchanged")
	}
}

func TestConcatPreHandler_Single(t *testing.T) {
	h := func(level Level, msg string, data M) M {
		data["modified"] = true
		return data
	}

	concat := ConcatPreHandler(h)
	result := concat(InfoLevel, "msg", M{})
	if result["modified"] != true {
		t.Error("single handler not applied")
	}
}
