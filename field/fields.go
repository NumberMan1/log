// Package field supplies structured fields without application-specific dependencies.
package field

// Field 表示一个调用方命名的结构化字段，不限定业务语义。
type Field struct {
	// Key 由调用方定义；重复键在合并时以后写入的值为准。
	Key string
	// Value 保留原始类型，复合值须按只读对象使用。
	Value any
}

// Fields 保存有序字段，合并时靠后的同名字段覆盖靠前的值。
type Fields []Field

// NewFields 复制传入的字段切片；字段中的复合值仍按只读对象使用。
func NewFields(fields ...Field) Fields { return append(Fields{}, fields...) }

// With 分配新切片后追加字段，保证兄弟 scope 不共享可写切片容量。
func (fields Fields) With(extra ...Field) Fields {
	result := make(Fields, 0, len(fields)+len(extra))
	result = append(result, fields...)
	return append(result, extra...)
}

// ToMap 生成独立顶层 map，同名字段取最后一个值。
func (fields Fields) ToMap() map[string]any {
	result := make(map[string]any, len(fields))
	for _, f := range fields {
		result[f.Key] = f.Value
	}
	return result
}

// Merge 将字段列表转换为独立 map，同名字段取最后一个值。
func Merge(fields ...Field) map[string]any { return Fields(fields).ToMap() }

// Any 创建任意值字段；不复制 map、指针等复合值。
func Any(key string, value any) Field { return Field{key, value} }

// Object 是 Any 的别名，不额外序列化或复制对象。
func Object(key string, value any) Field { return Any(key, value) }

// String 创建字符串字段。
func String(key, value string) Field { return Any(key, value) }

// Bool 创建布尔字段。
func Bool(key string, value bool) Field { return Any(key, value) }

// Int 创建 int 字段。
func Int(key string, value int) Field { return Any(key, value) }

// Int32 创建 int32 字段。
func Int32(key string, value int32) Field { return Any(key, value) }

// Int64 创建 int64 字段。
func Int64(key string, value int64) Field { return Any(key, value) }

// Uint64 创建 uint64 字段。
func Uint64(key string, value uint64) Field { return Any(key, value) }

// Float64 创建 float64 字段。
func Float64(key string, value float64) Field { return Any(key, value) }

// Strings 创建字符串切片字段，并复制输入切片。
func Strings(key string, value []string) Field { return Any(key, append([]string(nil), value...)) }

// Ints 创建 int 切片字段，并复制输入切片。
func Ints(key string, value []int) Field { return Any(key, append([]int(nil), value...)) }

// Int32s 创建 int32 切片字段，并复制输入切片。
func Int32s(key string, value []int32) Field { return Any(key, append([]int32(nil), value...)) }

// Int64s 创建 int64 切片字段，并复制输入切片。
func Int64s(key string, value []int64) Field { return Any(key, append([]int64(nil), value...)) }

// Error 以通用 error 键保留原始错误，不解包业务类型或丢失错误链。
func Error(err error) Field { return Any("error", err) }

// TraceID 创建调用方提供的 trace_id 字段，不生成或查询追踪上下文。
func TraceID(id string) Field { return String("trace_id", id) }

// SpanID 创建调用方提供的 span_id 字段，不生成追踪信息。
func SpanID(id string) Field { return String("span_id", id) }

// ParentSpanID 创建调用方提供的 parent_span_id 字段。
func ParentSpanID(id string) Field { return String("parent_span_id", id) }
