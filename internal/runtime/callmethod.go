package runtime

import (
	"fmt"
	"reflect"
)

func callMethod(recv Value, name string, args []Value) (Value, error) {
	if recv.kind != KindObject {
		return Nil(), fmt.Errorf("cannot call method %q on %s", name, kindName(recv.kind))
	}
	rv := reflect.ValueOf(recv.obj)
	m := rv.MethodByName(name)
	if !m.IsValid() {
		m = rv.MethodByName(exportedName(name))
	}
	if !m.IsValid() && rv.CanAddr() {
		m = rv.Addr().MethodByName(exportedName(name))
	}
	if !m.IsValid() {
		return Nil(), fmt.Errorf("method %q not found", name)
	}

	mt := m.Type()
	in := make([]reflect.Value, len(args))
	if mt.IsVariadic() {
		fixed := mt.NumIn() - 1
		for i := 0; i < len(args); i++ {
			var target reflect.Type
			if i < fixed {
				target = mt.In(i)
			} else {
				target = mt.In(fixed).Elem()
			}
			cv, err := coerceArg(args[i], target)
			if err != nil {
				return Nil(), err
			}
			in[i] = cv
		}
	} else {
		if len(args) != mt.NumIn() {
			return Nil(), fmt.Errorf("method %q expects %d arguments, got %d", name, mt.NumIn(), len(args))
		}
		for i := 0; i < len(args); i++ {
			cv, err := coerceArg(args[i], mt.In(i))
			if err != nil {
				return Nil(), err
			}
			in[i] = cv
		}
	}

	out := m.Call(in)
	if len(out) == 0 {
		return Nil(), nil
	}
	if len(out) == 2 {
		if err, ok := out[1].Interface().(error); ok && err != nil {
			return Nil(), err
		}
	}
	return FromAny(out[0].Interface()), nil
}

func coerceArg(v Value, target reflect.Type) (reflect.Value, error) {
	raw := v.Interface()
	if raw == nil {
		return reflect.Zero(target), nil
	}
	rv := reflect.ValueOf(raw)
	if rv.Type().AssignableTo(target) {
		return rv, nil
	}
	if rv.Type().ConvertibleTo(target) {
		return rv.Convert(target), nil
	}
	return reflect.Value{}, fmt.Errorf("cannot use %s as %s", rv.Type(), target)
}
