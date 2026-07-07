package runtime

import (
	"math"
	"strings"
	"testing"
)

func TestValueString(t *testing.T) {
	cases := []struct {
		v    Value
		want string
	}{
		{Nil(), ""},
		{Bool(true), "true"},
		{Bool(false), "false"},
		{Int(42), "42"},
		{Float(3.5), "3.5"},
		{Float(1024), "1024"},
		{Str("hola"), "hola"},
	}
	for _, tc := range cases {
		if got := tc.v.String(); got != tc.want {
			t.Errorf("String() = %q, want %q", got, tc.want)
		}
	}
}

func TestValueTruthy(t *testing.T) {
	cases := []struct {
		v    Value
		want bool
	}{
		{Nil(), false},
		{Bool(true), true},
		{Bool(false), false},
		{Int(0), false},
		{Int(1), true},
		{Str(""), false},
		{Str("x"), true},
		{Object([]Value{}), false},
		{Object([]Value{Int(1)}), true},
	}
	for _, tc := range cases {
		if got := tc.v.IsTruthy(); got != tc.want {
			t.Errorf("%v IsTruthy() = %v, want %v", tc.v.String(), got, tc.want)
		}
	}
}

func TestFromAny(t *testing.T) {
	if FromAny(int64(7)).Kind() != KindInt {
		t.Error("int64 should map to KindInt")
	}
	if FromAny("s").Kind() != KindString {
		t.Error("string should map to KindString")
	}
	if FromAny(nil).Kind() != KindNil {
		t.Error("nil should map to KindNil")
	}
	if FromAny(map[string]int{"a": 1}).Kind() != KindObject {
		t.Error("map should map to KindObject")
	}
}

func TestArithmetic(t *testing.T) {
	sum, err := arithAdd(Int(2), Int(3))
	if err != nil || sum.String() != "5" {
		t.Errorf("2+3 = %v (%v)", sum.String(), err)
	}
	div, err := arithDiv(Int(10), Int(4))
	if err != nil || div.String() != "2.5" {
		t.Errorf("10/4 = %v (%v)", div.String(), err)
	}
	if _, err := arithDiv(Int(1), Int(0)); err == nil {
		t.Error("expected division by zero error")
	}
}

func TestValueEqNumeric(t *testing.T) {
	cases := []struct {
		name string
		a, b Value
		want bool
	}{
		{"big_ints_distinct", Int(math.MaxInt64), Int(math.MaxInt64 - 1), false},
		{"big_ints_same", Int(math.MaxInt64), Int(math.MaxInt64), true},
		{"int_float_same", Int(3), Float(3.0), true},
		{"int_float_frac", Int(3), Float(3.5), false},
		{"int_float_overflow", Int(math.MaxInt64), Float(float64(math.MaxInt64)), false},
		{"bool_int", Bool(true), Int(1), true},
		{"bool_int_zero", Bool(false), Int(0), true},
		{"bool_int_other", Bool(true), Int(2), false},
		{"bool_float", Bool(true), Float(1.0), true},
	}
	for _, tc := range cases {
		if got := valueEq(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: valueEq = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestValueCmpBigInts(t *testing.T) {
	cmp, err := valueCmp(Int(math.MaxInt64), Int(math.MaxInt64-1))
	if err != nil || cmp != 1 {
		t.Errorf("cmp = %d (%v), want 1", cmp, err)
	}
	cmp, err = valueCmp(Int(math.MinInt64), Int(math.MinInt64+1))
	if err != nil || cmp != -1 {
		t.Errorf("cmp = %d (%v), want -1", cmp, err)
	}
}

func TestArithModFloat(t *testing.T) {
	res, err := arithMod(Float(10.5), Int(3))
	if err != nil || res.String() != "1.5" {
		t.Errorf("10.5 %% 3 = %v (%v), want 1.5", res.String(), err)
	}
	res, err = arithMod(Int(10), Int(3))
	if err != nil || res.Kind() != KindInt || res.String() != "1" {
		t.Errorf("10 %% 3 = %v (%v), want int 1", res.String(), err)
	}
	if _, err := arithMod(Int(1), Int(0)); err == nil {
		t.Error("expected modulo by zero error for ints")
	}
	if _, err := arithMod(Float(1.5), Float(0)); err == nil {
		t.Error("expected modulo by zero error for floats")
	}
}

func TestFromAnyUnsignedOverflow(t *testing.T) {
	v := FromAny(uint64(math.MaxUint64))
	if v.Kind() != KindFloat || v.AsFloat() <= 0 {
		t.Errorf("uint64 max = kind %d value %v, want positive float", v.Kind(), v.AsFloat())
	}
	v = FromAny(uint64(42))
	if v.Kind() != KindInt || v.AsInt() != 42 {
		t.Errorf("uint64 42 = kind %d value %d, want int 42", v.Kind(), v.AsInt())
	}
}

func TestLenRuneCount(t *testing.T) {
	if got := Str("año").Len(); got != 3 {
		t.Errorf("Str len = %d, want 3", got)
	}
	if got := Object(SafeString("año")).Len(); got != 3 {
		t.Errorf("SafeString len = %d, want 3", got)
	}
}

func TestDisassemble(t *testing.T) {
	p := &Program{
		Name:   "demo",
		Instrs: []Instr{{Op: OpLoadVar, Arg: 0}, {Op: OpEcho}},
		Consts: []Value{Str("name")},
	}
	dis := Disassemble(p)
	if !strings.Contains(dis, "LOAD_VAR") || !strings.Contains(dis, "ECHO") {
		t.Errorf("disassembly missing ops:\n%s", dis)
	}
}
