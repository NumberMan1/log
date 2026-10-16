package serverlog

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// SentryLogEvent 是映射器接收的原始日志快照，Data 中的 error 尚未字符串化。
type SentryLogEvent struct {
	// Level 是原始日志等级。
	Level Level
	// Message 是未经 Sentry 格式化的消息。
	Message string
	// Data 保留原始字段和 error；嵌套对象不可修改。
	Data M
	// Time 是提交事件的本地时间。
	Time time.Time
}

// SentryFields 描述 Sentry 原生字段及结构化附加数据。
// 映射器返回的嵌套对象应视为不可变，避免与 SDK 异步序列化竞争。
type SentryFields struct {
	// Tags 是需要索引的字符串标签；选择哪些字段索引由调用方决定。
	Tags map[string]string
	// User 是显式映射的 Sentry 用户信息，默认保持零值。
	User sentry.User
	// Extra 保留结构化附加字段；SDK v0.47 通过 Contexts["extra"] 承载。
	Extra M
	// Exception 是原生异常链和栈信息，不应再在 Extra 中重复存储同一 error。
	Exception []sentry.Exception
	// Contexts 是额外的命名上下文；extra 名称在 Extra 非空时由输出器使用。
	Contexts map[string]sentry.Context
}

// SentryFieldMapper 将通用日志映射到 Sentry，实现在并发 Log 中必须安全。
type SentryFieldMapper interface {
	// Map 创建目标字段快照，不修改事件中的共享嵌套对象。
	Map(SentryLogEvent) SentryFields
}

// SentryFieldMapperFunc 将普通函数适配成字段映射器。
type SentryFieldMapperFunc func(SentryLogEvent) SentryFields

// Map 执行字段映射；实现不应修改原始日志中的共享嵌套对象。
func (fn SentryFieldMapperFunc) Map(event SentryLogEvent) SentryFields { return fn(event) }

type defaultSentryFieldMapper struct {
	tags, userIDs, exceptions, excluded []string
	maxLength                           int
}

// NewDefaultSentryFieldMapper 创建通用映射器并复制所有规则切片。
// 默认只将 error/err 的原始错误转换为 Exception；tag、用户和排除规则均需显式配置。
// ExceptionKeys 为 nil 使用默认值，非 nil 空切片禁用异常映射。
func NewDefaultSentryFieldMapper(config SentryPrinterConfig) SentryFieldMapper {
	defaults := func(keys, fallback []string) []string {
		if keys == nil {
			keys = fallback
		}
		return append([]string{}, keys...)
	}
	length := config.MaxFieldLength
	if length <= 0 {
		length = 4096
	}
	return &defaultSentryFieldMapper{
		tags:       defaults(config.TagKeys, nil),
		userIDs:    defaults(config.UserIDKeys, nil),
		exceptions: defaults(config.ExceptionKeys, []string{"error", "err"}),
		excluded:   defaults(config.ExcludeKeys, nil),
		maxLength:  length,
	}
}

// Map 执行字段映射；实现不应修改原始日志中的共享嵌套对象。
func (mapper *defaultSentryFieldMapper) Map(event SentryLogEvent) SentryFields {
	fields := SentryFields{Tags: map[string]string{}, Extra: cloneData(event.Data)}
	if fields.Extra == nil {
		fields.Extra = M{}
	}
	for _, key := range mapper.excluded {
		delete(fields.Extra, key)
	}
	for _, key := range mapper.exceptions {
		if err, ok := fields.Extra[key].(error); ok && !isNilValue(err) {
			event := sentry.NewEvent()
			event.SetException(err, -1)
			for i := range event.Exception {
				if event.Exception[i].Stacktrace == nil {
					event.Exception[i].Stacktrace = sentry.NewStacktrace()
				}
			}
			fields.Exception = append(fields.Exception, event.Exception...)
			delete(fields.Extra, key)
		}
	}
	for _, key := range mapper.tags {
		if value, ok := fields.Extra[key]; ok && !isNilValue(value) {
			fields.Tags[key] = truncateUTF8(fmt.Sprint(value), mapper.maxLength)
			delete(fields.Extra, key)
		}
	}
	for _, key := range mapper.userIDs {
		value, ok := fields.Extra[key]
		if !ok || isNilValue(value) {
			continue
		}
		id := fmt.Sprint(value)
		if id == "" || id == "0" {
			continue
		}
		fields.User.ID = truncateUTF8(id, mapper.maxLength)
		delete(fields.Extra, key)
		break
	}
	for key, value := range fields.Extra {
		fields.Extra[key] = normalizeSentryValue(value, mapper.maxLength, 0)
	}
	return fields
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// Bound recursive payloads so cyclic maps and very deep structures remain safe.
func normalizeSentryValue(value any, maxLength, depth int) any {
	if isNilValue(value) {
		return nil
	}
	if depth >= 16 {
		return "[max depth]"
	}
	switch v := value.(type) {
	case string:
		return truncateUTF8(v, maxLength)
	case error:
		return truncateUTF8(v.Error(), maxLength)
	case fmt.Stringer:
		return truncateUTF8(v.String(), maxLength)
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return truncateUTF8(fmt.Sprint(value), maxLength)
		}
		result := M{}
		iter := v.MapRange()
		for count := 0; iter.Next() && count < 1000; count++ {
			result[iter.Key().String()] = normalizeSentryValue(iter.Value().Interface(), maxLength, depth+1)
		}
		return result
	case reflect.Slice, reflect.Array:
		length := min(v.Len(), 1000)
		result := make([]any, length)
		for i := range result {
			result[i] = normalizeSentryValue(v.Index(i).Interface(), maxLength, depth+1)
		}
		return result
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return value
	default:
		return truncateUTF8(fmt.Sprint(value), maxLength)
	}
}

func truncateUTF8(value string, maxLength int) string {
	value = strings.ToValidUTF8(value, "�")
	if len(value) <= maxLength {
		return value
	}
	end := maxLength
	for end > 0 && !utf8.RuneStart(value[end]) {
		end--
	}
	return value[:end]
}
