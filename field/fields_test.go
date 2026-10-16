package field

import (
	"errors"
	"testing"
)

func TestFieldsScopeAndOverride(t *testing.T) {
	parent := make(Fields, 1, 8)
	parent[0] = String("scope", "root")
	a := parent.With(String("child", "a"))
	b := parent.With(String("child", "b"))
	if a.ToMap()["child"] != "a" || b.ToMap()["child"] != "b" || len(parent) != 1 {
		t.Fatal("sibling scopes share backing array")
	}
	merged := NewFields(String("key", "old")).With(String("key", "new")).ToMap()
	if merged["key"] != "new" {
		t.Fatal("call fields must override inherited fields")
	}
}

func TestErrorRetainsIdentityAndSlicesAreCopied(t *testing.T) {
	err := errors.New("original")
	if Error(err).Value != err || Error(nil).Value != nil {
		t.Fatal("error identity lost")
	}
	values := []string{"original"}
	f := Strings("strings", values)
	values[0] = "modified"
	if f.Value.([]string)[0] != "original" {
		t.Fatal("slice not copied")
	}
	fields := Merge(Any("any", 1), Object("object", true), Bool("bool", true), Int("int", 1), Int32("int32", 2), Int64("int64", 3), Uint64("uint", 4), Float64("float", 1.5), Ints("ints", []int{1}), Int32s("int32s", []int32{2}), Int64s("int64s", []int64{3}), TraceID("t"), SpanID("s"), ParentSpanID("p"))
	if len(fields) != 14 || fields["trace_id"] != "t" {
		t.Fatal(fields)
	}
}
