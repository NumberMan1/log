package field

import (
	"encoding/json"
	"github.com/NumberMan1/general/log/field"
	"go.uber.org/zap"
)

type Field = field.Field

type Fields = field.Fields

var NewFields = field.NewFields

// 有需要新的Field可以在这里引入
var (
	String    = zap.String
	Strings   = zap.Strings
	Bool      = zap.Bool
	Error     = zap.Error
	StackSkip = zap.StackSkip
	AnyString = func(key string, value interface{}) Field {
		valData, _ := json.Marshal(value)
		return zap.String(key, string(valData))
	}

	Any = zap.Any

	Int      = zap.Int
	Int8     = zap.Int8
	Int16    = zap.Int16
	Int32    = zap.Int32
	Int64    = zap.Int64
	Int64s   = zap.Int64s
	Uint8    = zap.Uint8
	Uint32   = zap.Uint32
	Uint64   = zap.Uint64
	Duration = zap.Duration
)
