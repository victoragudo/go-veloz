package runtime

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
)

var errModuloByZero = errors.New("modulo by zero")

func kindName(k Kind) string {
	switch k {
	case KindNil:
		return "null"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindString:
		return "string"
	default:
		return "object"
	}
}

func opErr(op string, a, b Value) error {
	return fmt.Errorf("unsupported operands for %q: %s and %s", op, kindName(a.kind), kindName(b.kind))
}

func arithAdd(a, b Value) (Value, error) {
	if a.kind == KindInt && b.kind == KindInt {
		return Int(a.AsInt() + b.AsInt()), nil
	}
	fa, oka := a.toFloat()
	fb, okb := b.toFloat()
	if oka && okb {
		return Float(fa + fb), nil
	}
	return Nil(), opErr("+", a, b)
}

func arithSub(a, b Value) (Value, error) {
	if a.kind == KindInt && b.kind == KindInt {
		return Int(a.AsInt() - b.AsInt()), nil
	}
	fa, oka := a.toFloat()
	fb, okb := b.toFloat()
	if oka && okb {
		return Float(fa - fb), nil
	}
	return Nil(), opErr("-", a, b)
}

func arithMul(a, b Value) (Value, error) {
	if a.kind == KindInt && b.kind == KindInt {
		return Int(a.AsInt() * b.AsInt()), nil
	}
	fa, oka := a.toFloat()
	fb, okb := b.toFloat()
	if oka && okb {
		return Float(fa * fb), nil
	}
	return Nil(), opErr("*", a, b)
}

func arithDiv(a, b Value) (Value, error) {
	fa, oka := a.toFloat()
	fb, okb := b.toFloat()
	if !oka || !okb {
		return Nil(), opErr("/", a, b)
	}
	if fb == 0 {
		return Nil(), errors.New("division by zero")
	}
	return Float(fa / fb), nil
}

func arithMod(a, b Value) (Value, error) {
	if a.kind == KindInt && b.kind == KindInt {
		ib := b.AsInt()
		if ib == 0 {
			return Nil(), errModuloByZero
		}
		return Int(a.AsInt() % ib), nil
	}
	fa, oka := a.toFloat()
	fb, okb := b.toFloat()
	if !oka || !okb {
		return Nil(), opErr("%", a, b)
	}
	if fb == 0 {
		return Nil(), errModuloByZero
	}
	return Float(math.Mod(fa, fb)), nil
}

func arithPow(a, b Value) (Value, error) {
	fa, oka := a.toFloat()
	fb, okb := b.toFloat()
	if !oka || !okb {
		return Nil(), opErr("**", a, b)
	}
	return Float(math.Pow(fa, fb)), nil
}

func arithNeg(a Value) (Value, error) {
	switch a.kind {
	case KindInt:
		return Int(-a.AsInt()), nil
	case KindFloat:
		return Float(-a.AsFloat()), nil
	default:
		if f, ok := a.toFloat(); ok {
			return Float(-f), nil
		}
	}
	return Nil(), fmt.Errorf("cannot negate %s", kindName(a.kind))
}

func valueEq(a, b Value) bool {
	if a.isNumericOrBool() && b.isNumericOrBool() {
		return numericEq(a, b)
	}
	if a.kind != b.kind {
		return a.kind == KindNil && b.kind == KindNil
	}
	switch a.kind {
	case KindString:
		return a.str == b.str
	case KindNil:
		return true
	case KindObject:
		return reflect.DeepEqual(a.obj, b.obj)
	}
	return false
}

func numericEq(a, b Value) bool {
	ia, aIsInt := a.exactInt()
	ib, bIsInt := b.exactInt()
	switch {
	case aIsInt && bIsInt:
		return ia == ib
	case aIsInt:
		return floatEqualsInt(b.AsFloat(), ia)
	case bIsInt:
		return floatEqualsInt(a.AsFloat(), ib)
	default:
		return a.AsFloat() == b.AsFloat()
	}
}

func floatEqualsInt(f float64, i int64) bool {
	return f >= math.MinInt64 && f < math.MaxInt64 && f == math.Trunc(f) && int64(f) == i
}

func valueCmp(a, b Value) (int, error) {
	if a.isNumericOrBool() && b.isNumericOrBool() {
		if ia, aIsInt := a.exactInt(); aIsInt {
			if ib, bIsInt := b.exactInt(); bIsInt {
				return cmpInt(ia, ib), nil
			}
		}
		fa, _ := a.toFloat()
		fb, _ := b.toFloat()
		switch {
		case fa < fb:
			return -1, nil
		case fa > fb:
			return 1, nil
		default:
			return 0, nil
		}
	}
	if a.kind == KindString && b.kind == KindString {
		return strings.Compare(a.str, b.str), nil
	}
	return 0, fmt.Errorf("cannot compare %s and %s", kindName(a.kind), kindName(b.kind))
}

func cmpInt(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func contains(item, seq Value) bool {
	switch seq.kind {
	case KindString:
		if item.kind == KindString {
			return strings.Contains(seq.str, item.str)
		}
		return strings.Contains(seq.str, item.String())
	case KindObject:
		rv := reflect.ValueOf(seq.obj)
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				return false
			}
			rv = rv.Elem()
		}
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < rv.Len(); i++ {
				if valueEq(item, FromAny(rv.Index(i).Interface())) {
					return true
				}
			}
		case reflect.Map:
			kt := rv.Type().Key()
			kv := reflect.ValueOf(item.Interface())
			if kv.IsValid() && kv.Type().ConvertibleTo(kt) {
				if rv.MapIndex(kv.Convert(kt)).IsValid() {
					return true
				}
			}
		}
	}
	return false
}
