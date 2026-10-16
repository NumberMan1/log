package serverlog

import (
	"encoding/json"
	"errors"
	"testing"
)

var jsonBenchmarkResult string

func jsonBenchmarkInputs() []struct {
	name string
	data M
} {
	return []struct {
		name string
		data M
	}{
		{"Flat6", M{"level": "info", "msg": "request completed", "operation": "fetch", "duration_ms": 12, "success": true, "count": 42}},
		{"Nested", M{
			"level": "info", "msg": "request completed", "operation": "fetch", "duration_ms": 12, "success": true,
			"metadata": M{"attempt": 2, "labels": []any{"one", "two", "three"}, "items": []any{M{"id": 1, "name": "first"}, M{"id": 2, "name": "second"}}},
		}},
	}
}

// DirectMarshal 使用不含 error 的相同输入作为参照，不提供错误转换或深度限制。
func BenchmarkJsonFormat(b *testing.B) {
	format := NewJsonFormat()
	for _, input := range jsonBenchmarkInputs() {
		b.Run(input.name+"/Format", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				jsonBenchmarkResult = format.Format(InfoLevel, "request completed", input.data)
			}
		})
		b.Run(input.name+"/DirectMarshal", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				out, err := json.Marshal(input.data)
				if err != nil {
					b.Fatal(err)
				}
				jsonBenchmarkResult = string(out)
			}
		})
	}
	b.Run("NestedError", func(b *testing.B) {
		data := M{"metadata": M{"attempt": 2, "items": []any{M{"id": 1}, errors.New("unavailable")}}, "success": false}
		b.ReportAllocs()
		for b.Loop() {
			jsonBenchmarkResult = format.Format(ErrorLevel, "failed", data)
		}
	})
}
