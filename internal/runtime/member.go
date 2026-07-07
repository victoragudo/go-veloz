package runtime

import (
	"reflect"
	"strings"
	"sync"
)

type Attributer interface {
	VelozAttr(name string) (Value, bool)
}

type accessorKey struct {
	typ  reflect.Type
	name string
}

type accessor func(reflect.Value) (Value, bool)

var accessorCache sync.Map

func resolveAttr(v Value, name string) Value {
	switch v.kind {
	case KindObject:
		if a, ok := v.obj.(Attributer); ok {
			if out, found := a.VelozAttr(name); found {
				return out
			}
			return Nil()
		}
		return reflectAttr(reflect.ValueOf(v.obj), name)
	default:
		return Nil()
	}
}

func reflectAttr(rv reflect.Value, name string) Value {
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return Nil()
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Map {
		return mapLookup(rv, name)
	}

	if rv.Kind() != reflect.Struct {
		return methodCall(rv, name)
	}

	key := accessorKey{typ: rv.Type(), name: name}
	if cached, ok := accessorCache.Load(key); ok {
		if out, found := cached.(accessor)(rv); found {
			return out
		}
		return Nil()
	}

	acc := buildAccessor(rv.Type(), name)
	accessorCache.Store(key, acc)
	if out, found := acc(rv); found {
		return out
	}
	return Nil()
}

func buildAccessor(typ reflect.Type, name string) accessor {
	if field, ok := typ.FieldByName(name); ok && field.PkgPath == "" {
		idx := field.Index
		return func(rv reflect.Value) (Value, bool) {
			return FromAny(rv.FieldByIndex(idx).Interface()), true
		}
	}
	title := exportedName(name)
	if title != name {
		if field, ok := typ.FieldByName(title); ok && field.PkgPath == "" {
			idx := field.Index
			return func(rv reflect.Value) (Value, bool) {
				return FromAny(rv.FieldByIndex(idx).Interface()), true
			}
		}
	}
	for _, candidate := range []string{name, title, "Get" + title} {
		if m, ok := methodByName(typ, candidate); ok {
			mIndex := m.Index
			return func(rv reflect.Value) (Value, bool) {
				return callNoArg(rv.Method(mIndex))
			}
		}
		if m, ok := methodByName(reflect.PtrTo(typ), candidate); ok {
			mIndex := m.Index
			return func(rv reflect.Value) (Value, bool) {
				if rv.CanAddr() {
					return callNoArg(rv.Addr().Method(mIndex))
				}
				pv := reflect.New(rv.Type())
				pv.Elem().Set(rv)
				return callNoArg(pv.Method(mIndex))
			}
		}
	}
	return func(reflect.Value) (Value, bool) { return Nil(), false }
}

func methodByName(typ reflect.Type, name string) (reflect.Method, bool) {
	m, ok := typ.MethodByName(name)
	if !ok {
		return reflect.Method{}, false
	}
	mt := m.Type
	in := mt.NumIn()
	if typ.Kind() == reflect.Ptr {
		in--
	}
	if in != 0 {
		return reflect.Method{}, false
	}
	if mt.NumOut() == 0 || mt.NumOut() > 2 {
		return reflect.Method{}, false
	}
	return m, true
}

func callNoArg(m reflect.Value) (Value, bool) {
	out := m.Call(nil)
	if len(out) == 2 {
		if err, ok := out[1].Interface().(error); ok && err != nil {
			return Nil(), false
		}
	}
	return FromAny(out[0].Interface()), true
}

func methodCall(rv reflect.Value, name string) Value {
	if !rv.IsValid() {
		return Nil()
	}
	if m, ok := methodByName(rv.Type(), name); ok {
		if out, found := callNoArg(rv.Method(m.Index)); found {
			return out
		}
	}
	if m, ok := methodByName(rv.Type(), exportedName(name)); ok {
		if out, found := callNoArg(rv.Method(m.Index)); found {
			return out
		}
	}
	return Nil()
}

func mapLookup(rv reflect.Value, name string) Value {
	kt := rv.Type().Key()
	if kt.Kind() == reflect.String {
		val := rv.MapIndex(reflect.ValueOf(name).Convert(kt))
		if !val.IsValid() {
			return Nil()
		}
		return FromAny(val.Interface())
	}
	return Nil()
}

func exportedName(name string) string {
	if name == "" {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func resolveIndex(v Value, idx Value) Value {
	if v.kind != KindObject {
		return Nil()
	}
	rv := reflect.ValueOf(v.obj)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return Nil()
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		i, ok := idx.toInt()
		if !ok {
			return Nil()
		}
		n := rv.Len()
		if i < 0 {
			i += int64(n)
		}
		if i < 0 || i >= int64(n) {
			return Nil()
		}
		return FromAny(rv.Index(int(i)).Interface())
	case reflect.Map:
		kt := rv.Type().Key()
		kv := reflect.ValueOf(idx.Interface())
		if !kv.IsValid() {
			return Nil()
		}
		if !kv.Type().ConvertibleTo(kt) {
			return Nil()
		}
		val := rv.MapIndex(kv.Convert(kt))
		if !val.IsValid() {
			return Nil()
		}
		return FromAny(val.Interface())
	case reflect.Struct:
		if idx.kind == KindString {
			return reflectAttr(rv, idx.str)
		}
	}
	return Nil()
}

func (v Value) toInt() (int64, bool) {
	switch v.kind {
	case KindInt:
		return int64(v.num), true
	case KindFloat:
		f := v.AsFloat()
		return int64(f), true
	case KindBool:
		if v.num != 0 {
			return 1, true
		}
		return 0, true
	case KindString:
		f, ok := v.toFloat()
		if !ok {
			return 0, false
		}
		return int64(f), true
	}
	return 0, false
}
