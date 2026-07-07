package runtime

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Kind uint8

const (
	KindNil Kind = iota
	KindBool
	KindInt
	KindFloat
	KindString
	KindObject
)

type Value struct {
	kind Kind
	num  uint64
	str  string
	obj  any
}

func Nil() Value { return Value{kind: KindNil} }

func Bool(b bool) Value {
	var n uint64
	if b {
		n = 1
	}
	return Value{kind: KindBool, num: n}
}

func Int(i int64) Value { return Value{kind: KindInt, num: uint64(i)} }

func Float(f float64) Value { return Value{kind: KindFloat, num: math.Float64bits(f)} }

func Str(s string) Value { return Value{kind: KindString, str: s} }

func Object(o any) Value {
	if o == nil {
		return Value{kind: KindNil}
	}
	return Value{kind: KindObject, obj: o}
}

func FromAny(o any) Value {
	switch x := o.(type) {
	case nil:
		return Value{kind: KindNil}
	case Value:
		return x
	case bool:
		return Bool(x)
	case string:
		return Str(x)
	case int:
		return Int(int64(x))
	case int64:
		return Int(x)
	case int32:
		return Int(int64(x))
	case int16:
		return Int(int64(x))
	case int8:
		return Int(int64(x))
	case uint:
		return fromUint64(uint64(x))
	case uint64:
		return fromUint64(x)
	case uint32:
		return Int(int64(x))
	case uint16:
		return Int(int64(x))
	case uint8:
		return Int(int64(x))
	case float64:
		return Float(x)
	case float32:
		return Float(float64(x))
	default:
		return Value{kind: KindObject, obj: o}
	}
}

func fromUint64(x uint64) Value {
	if x > math.MaxInt64 {
		return Float(float64(x))
	}
	return Int(int64(x))
}

func (v Value) Kind() Kind { return v.kind }

func (v Value) IsNil() bool { return v.kind == KindNil }

func (v Value) AsBool() bool { return v.num != 0 }

func (v Value) AsInt() int64 { return int64(v.num) }

func (v Value) AsFloat() float64 { return math.Float64frombits(v.num) }

func (v Value) Interface() any {
	switch v.kind {
	case KindNil:
		return nil
	case KindBool:
		return v.num != 0
	case KindInt:
		return int64(v.num)
	case KindFloat:
		return math.Float64frombits(v.num)
	case KindString:
		return v.str
	default:
		return v.obj
	}
}

func (v Value) IsTruthy() bool {
	switch v.kind {
	case KindNil:
		return false
	case KindBool:
		return v.num != 0
	case KindInt:
		return int64(v.num) != 0
	case KindFloat:
		return math.Float64frombits(v.num) != 0
	case KindString:
		return v.str != ""
	case KindObject:
		return objectTruthy(v.obj)
	}
	return false
}

func objectTruthy(o any) bool {
	if o == nil {
		return false
	}
	if s, ok := o.(SafeString); ok {
		return s != ""
	}
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil()
	}
	return true
}

func (v Value) String() string {
	switch v.kind {
	case KindNil:
		return ""
	case KindBool:
		if v.num != 0 {
			return "true"
		}
		return "false"
	case KindInt:
		return strconv.FormatInt(int64(v.num), 10)
	case KindFloat:
		return formatFloat(math.Float64frombits(v.num))
	case KindString:
		return v.str
	case KindObject:
		return objectString(v.obj)
	}
	return ""
}

func objectString(o any) string {
	switch x := o.(type) {
	case SafeString:
		return string(x)
	case string:
		return x
	case []byte:
		return string(x)
	case fmt.Stringer:
		return x.String()
	case error:
		return x.Error()
	default:
		return fmt.Sprint(o)
	}
}

func formatFloat(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (v Value) isNumericOrBool() bool {
	return v.kind == KindInt || v.kind == KindFloat || v.kind == KindBool
}

func (v Value) exactInt() (int64, bool) {
	switch v.kind {
	case KindInt:
		return int64(v.num), true
	case KindBool:
		if v.num != 0 {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func (v Value) toFloat() (float64, bool) {
	switch v.kind {
	case KindInt:
		return float64(int64(v.num)), true
	case KindFloat:
		return math.Float64frombits(v.num), true
	case KindBool:
		if v.num != 0 {
			return 1, true
		}
		return 0, true
	case KindString:
		f, err := strconv.ParseFloat(strings.TrimSpace(v.str), 64)
		return f, err == nil
	}
	return 0, false
}

func (v Value) Len() int {
	switch v.kind {
	case KindString:
		return utf8.RuneCountInString(v.str)
	case KindObject:
		rv := reflect.ValueOf(v.obj)
		switch rv.Kind() {
		case reflect.String:
			return utf8.RuneCountInString(rv.String())
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len()
		}
	}
	return 0
}

func (v Value) isCollection() bool {
	if v.kind != KindObject {
		return false
	}
	rv := reflect.ValueOf(v.obj)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	}
	return false
}
