package veloz

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"veloz/internal/runtime"
)

func registerFilters(e *Engine) {
	f := map[string]Callable{
		"upper":      filterUpper,
		"lower":      filterLower,
		"capitalize": filterCapitalize,
		"title":      filterTitle,
		"trim":       filterTrim,
		"length":     filterLength,
		"count":      filterLength,
		"default":    filterDefault,
		"join":       filterJoin,
		"reverse":    filterReverse,
		"first":      filterFirst,
		"last":       filterLast,
		"keys":       filterKeys,
		"abs":        filterAbs,
		"round":      filterRound,
		"replace":    filterReplace,
		"split":      filterSplit,
		"escape":     filterEscape,
		"e":          filterEscape,
		"raw":        filterRaw,
		"nl2br":      filterNl2br,
	}
	for name, fn := range f {
		e.filters[name] = fn
	}
}

func asFloat(v Value) (float64, bool) {
	switch x := v.Interface().(type) {
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f, err == nil
	}
	return 0, false
}

func asInt(v Value) (int64, bool) {
	switch x := v.Interface().(type) {
	case int64:
		return x, true
	case float64:
		return int64(x), true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		if err == nil {
			return i, true
		}
		f, err2 := strconv.ParseFloat(strings.TrimSpace(x), 64)
		if err2 == nil {
			return int64(f), true
		}
	}
	return 0, false
}

func toList(v Value) []Value {
	raw := v.Interface()
	if raw == nil {
		return nil
	}
	rv := reflect.ValueOf(raw)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		out := make([]Value, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out[i] = FromAny(rv.Index(i).Interface())
		}
		return out
	}
	return nil
}

func filterUpper(args []Value) (Value, error) {
	return Str(strings.ToUpper(args[0].String())), nil
}

func filterLower(args []Value) (Value, error) {
	return Str(strings.ToLower(args[0].String())), nil
}

func filterCapitalize(args []Value) (Value, error) {
	s := args[0].String()
	if s == "" {
		return Str(s), nil
	}
	r := []rune(s)
	out := make([]rune, len(r))
	out[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		out[i] = unicode.ToLower(r[i])
	}
	return Str(string(out)), nil
}

func filterTitle(args []Value) (Value, error) {
	var b strings.Builder
	atWordStart := true
	for _, r := range args[0].String() {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if atWordStart {
				b.WriteRune(unicode.ToUpper(r))
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
			atWordStart = false
		} else {
			b.WriteRune(r)
			atWordStart = true
		}
	}
	return Str(b.String()), nil
}

func filterTrim(args []Value) (Value, error) {
	if len(args) >= 2 {
		return Str(strings.Trim(args[0].String(), args[1].String())), nil
	}
	return Str(strings.TrimSpace(args[0].String())), nil
}

func filterLength(args []Value) (Value, error) {
	return Int(int64(args[0].Len())), nil
}

func filterDefault(args []Value) (Value, error) {
	if len(args) >= 2 && !args[0].IsTruthy() {
		return args[1], nil
	}
	return args[0], nil
}

func filterJoin(args []Value) (Value, error) {
	sep := ""
	if len(args) >= 2 {
		sep = args[1].String()
	}
	list := toList(args[0])
	parts := make([]string, len(list))
	for i, e := range list {
		parts[i] = e.String()
	}
	return Str(strings.Join(parts, sep)), nil
}

func filterReverse(args []Value) (Value, error) {
	v := args[0]
	if v.Kind() == runtime.KindString {
		r := []rune(v.String())
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return Str(string(r)), nil
	}
	list := toList(v)
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return Object(list), nil
}

func filterFirst(args []Value) (Value, error) {
	v := args[0]
	if v.Kind() == runtime.KindString {
		r := []rune(v.String())
		if len(r) == 0 {
			return Str(""), nil
		}
		return Str(string(r[0])), nil
	}
	list := toList(v)
	if len(list) == 0 {
		return Nil(), nil
	}
	return list[0], nil
}

func filterLast(args []Value) (Value, error) {
	v := args[0]
	if v.Kind() == runtime.KindString {
		r := []rune(v.String())
		if len(r) == 0 {
			return Str(""), nil
		}
		return Str(string(r[len(r)-1])), nil
	}
	list := toList(v)
	if len(list) == 0 {
		return Nil(), nil
	}
	return list[len(list)-1], nil
}

func filterKeys(args []Value) (Value, error) {
	raw := args[0].Interface()
	if raw == nil {
		return Object([]Value{}), nil
	}
	rv := reflect.ValueOf(raw)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return Object([]Value{}), nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Map:
		keys := rv.MapKeys()
		sort.Slice(keys, func(a, b int) bool {
			return fmt.Sprint(keys[a].Interface()) < fmt.Sprint(keys[b].Interface())
		})
		out := make([]Value, len(keys))
		for i, k := range keys {
			out[i] = FromAny(k.Interface())
		}
		return Object(out), nil
	case reflect.Slice, reflect.Array:
		out := make([]Value, rv.Len())
		for i := range out {
			out[i] = Int(int64(i))
		}
		return Object(out), nil
	}
	return Object([]Value{}), nil
}

func filterAbs(args []Value) (Value, error) {
	v := args[0]
	if v.Kind() == runtime.KindInt {
		i := v.AsInt()
		if i < 0 {
			i = -i
		}
		return Int(i), nil
	}
	f, ok := asFloat(v)
	if !ok {
		return Nil(), fmt.Errorf("abs: %s is not a number", v.String())
	}
	return Float(math.Abs(f)), nil
}

func filterRound(args []Value) (Value, error) {
	f, ok := asFloat(args[0])
	if !ok {
		return Nil(), fmt.Errorf("round: %s is not a number", args[0].String())
	}
	prec := 0
	if len(args) >= 2 {
		if p, ok := asInt(args[1]); ok {
			prec = int(p)
		}
	}
	mult := math.Pow(10, float64(prec))
	return Float(math.Round(f*mult) / mult), nil
}

func filterReplace(args []Value) (Value, error) {
	s := args[0].String()
	if len(args) == 2 {
		if m, ok := args[1].Interface().(map[string]Value); ok {
			for k, v := range m {
				s = strings.ReplaceAll(s, k, v.String())
			}
			return Str(s), nil
		}
	}
	if len(args) >= 3 {
		return Str(strings.ReplaceAll(s, args[1].String(), args[2].String())), nil
	}
	return Str(s), nil
}

func filterSplit(args []Value) (Value, error) {
	sep := ""
	if len(args) >= 2 {
		sep = args[1].String()
	}
	parts := strings.Split(args[0].String(), sep)
	out := make([]Value, len(parts))
	for i, p := range parts {
		out[i] = Str(p)
	}
	return Object(out), nil
}

func filterEscape(args []Value) (Value, error) {
	return Object(SafeString(html.EscapeString(args[0].String()))), nil
}

func filterRaw(args []Value) (Value, error) {
	return Object(SafeString(args[0].String())), nil
}

func filterNl2br(args []Value) (Value, error) {
	s := html.EscapeString(args[0].String())
	s = strings.ReplaceAll(s, "\n", "<br />\n")
	return Object(SafeString(s)), nil
}
