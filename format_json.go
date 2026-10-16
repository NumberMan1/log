package serverlog

import (
	"encoding/json"
	"reflect"
)

const jsonLogMaxDepth = 32

// JsonFormat 仅将 data 序列化为 JSON，不隐式加入 msg 或 level。
type JsonFormat struct {
}

// NewJsonFormat 创建可并发使用的 JSON 格式器。
func NewJsonFormat() *JsonFormat {
	return &JsonFormat{}
}

// Format 在序列化边界将嵌套 error 转为字符串，不修改输入。
// data 为 nil 返回空串；序列化失败时返回带错误说明的有效 JSON。
// 普通 map/slice 仅在需要转换子项时复制；调用期间调用方不得修改输入及嵌套对象。
func (obj *JsonFormat) Format(level Level, msg string, data M) (result string) {
	if data == nil {
		return ""
	}

	var seqBytes []byte
	var err error
	value, _ := jsonLogValue(data, 0)
	seqBytes, err = json.Marshal(value)
	if err != nil {
		seqBytes, _ = json.Marshal(map[string]string{"error": err.Error(), "msg": msg, "data": "[unserializable]"})
	}
	return string(seqBytes)
}

// jsonLogValue 返回转换结果及是否替换了原值，避免比较可能不可比较的动态类型。
// 常见类型不经过反射；map/slice 延迟到首个变化子项时才复制，未变化的分支只读复用。
// 深度限制也适用于快路径，防止循环容器无限递归；命名类型保留通用反射处理。
func jsonLogValue(value any, depth int) (any, bool) {
	if value == nil {
		return nil, false
	}
	if depth >= jsonLogMaxDepth {
		if isNilValue(value) {
			return nil, true
		}
		return "[max depth]", true
	}
	switch v := value.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, json.Number, []byte:
		return value, false
	case []string, []bool,
		[]int, []int8, []int16, []int32, []int64,
		[]uint, []uint16, []uint32, []uint64, []uintptr,
		[]float32, []float64, []json.Number:
		// 元素不可能包含 error；仅在元素触及深度上限时才逐项转换。
		if depth < jsonLogMaxDepth-1 {
			return value, false
		}
	case map[string]any:
		var result M
		for key, item := range v {
			converted, changed := jsonLogValue(item, depth+1)
			if changed {
				if result == nil {
					result = cloneData(v)
				}
				result[key] = converted
			}
		}
		if result != nil {
			return result, true
		}
		return value, false
	case []any:
		var result []any
		for i, item := range v {
			converted, changed := jsonLogValue(item, depth+1)
			if changed {
				if result == nil {
					result = append([]any(nil), v...)
				}
				result[i] = converted
			}
		}
		if result != nil {
			return result, true
		}
		return value, false
	}
	// 必须在调用 error/Marshaler 方法之前识别 typed nil。
	if isNilValue(value) {
		return nil, true
	}
	if err, ok := value.(error); ok {
		return err.Error(), true
	}
	if _, ok := value.(json.Marshaler); ok {
		return value, false
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return value, false
		}
		result := make(M, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			converted, _ := jsonLogValue(iter.Value().Interface(), depth+1)
			result[iter.Key().String()] = converted
		}
		return result, true
	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := range result {
			result[i], _ = jsonLogValue(v.Index(i).Interface(), depth+1)
		}
		return result, true
	default:
		return value, false
	}
}
