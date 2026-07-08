package veloz

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/victoragudo/go-veloz/internal/runtime"
)

func requireArg(name string, fn Callable) Callable {
	return func(args []Value) (Value, error) {
		if len(args) == 0 {
			return Nil(), fmt.Errorf("%s: missing argument", name)
		}
		return fn(args)
	}
}

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
		"truncate":   filterTruncate,
		"slice":      filterSlice,
		"batch":      filterBatch,
		"sort":       filterSort,
		"date":       filterDate,
		"map":        e.filterMap,
	}
	for name, fn := range f {
		e.filters[name] = requireArg(name, fn)
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
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
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
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
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

const defaultTruncateLen = 255

func filterTruncate(args []Value) (Value, error) {
	s := args[0].String()
	length := int64(defaultTruncateLen)
	if len(args) >= 2 {
		if n, ok := asInt(args[1]); ok && n >= 0 {
			length = n
		}
	}
	suffix := "..."
	if len(args) >= 3 {
		suffix = args[2].String()
	}
	r := []rune(s)
	if int64(len(r)) <= length {
		return Str(s), nil
	}
	return Str(string(r[:length]) + suffix), nil
}

func filterSlice(args []Value) (Value, error) {
	if len(args) < 2 {
		return Nil(), fmt.Errorf("slice: expects a start index")
	}
	start, ok := asInt(args[1])
	if !ok {
		return Nil(), fmt.Errorf("slice: start must be numeric")
	}
	sliceBounds := func(size int64) (int64, int64) {
		if start < 0 {
			start += size
		}
		start = min(max(start, 0), size)
		end := size
		if len(args) >= 3 {
			if length, lok := asInt(args[2]); lok && length >= 0 {
				end = min(start+length, size)
			}
		}
		return start, end
	}
	if args[0].Kind() == runtime.KindString {
		r := []rune(args[0].String())
		from, to := sliceBounds(int64(len(r)))
		return Str(string(r[from:to])), nil
	}
	list := toList(args[0])
	from, to := sliceBounds(int64(len(list)))
	return Object(list[from:to]), nil
}

func filterBatch(args []Value) (Value, error) {
	if len(args) < 2 {
		return Nil(), fmt.Errorf("batch: expects a group size")
	}
	size, ok := asInt(args[1])
	if !ok || size <= 0 {
		return Nil(), fmt.Errorf("batch: size must be a positive number")
	}
	list := toList(args[0])
	out := make([]Value, 0, (int64(len(list))+size-1)/size)
	for from := int64(0); from < int64(len(list)); from += size {
		to := min(from+size, int64(len(list)))
		group := make([]Value, to-from, size)
		copy(group, list[from:to])
		if len(args) >= 3 {
			for int64(len(group)) < size {
				group = append(group, args[2])
			}
		}
		out = append(out, Object(group))
	}
	return Object(out), nil
}

func filterSort(args []Value) (Value, error) {
	list := toList(args[0])
	attr := ""
	if len(args) >= 2 {
		attr = args[1].String()
	}
	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if attr != "" {
			a, b = a.Attr(attr), b.Attr(attr)
		}
		return compareValues(a, b) < 0
	})
	return Object(list), nil
}

const defaultDateLayout = "2006-01-02 15:04:05"

func filterDate(args []Value) (Value, error) {
	t, ok := toTime(args[0])
	if !ok {
		return Nil(), fmt.Errorf("date: cannot interpret %q as a time", args[0].String())
	}
	layout := defaultDateLayout
	if len(args) >= 2 {
		layout = args[1].String()
	}
	return Str(t.Format(layout)), nil
}

func toTime(v Value) (time.Time, bool) {
	switch x := v.Interface().(type) {
	case time.Time:
		return x, true
	case *time.Time:
		if x != nil {
			return *x, true
		}
	case int64:
		return time.Unix(x, 0).UTC(), true
	case float64:
		return time.Unix(int64(x), 0).UTC(), true
	case string:
		for _, layout := range []string{time.RFC3339, defaultDateLayout, "2006-01-02"} {
			if t, err := time.Parse(layout, x); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func (e *Engine) filterMap(args []Value) (Value, error) {
	if len(args) < 2 {
		return Nil(), fmt.Errorf("map: expects a filter or attribute name")
	}
	list := toList(args[0])
	name := args[1].String()
	out := make([]Value, len(list))
	if fn, ok := e.ResolveCallable(name, true); ok {
		for i, item := range list {
			v, err := fn(append([]Value{item}, args[2:]...))
			if err != nil {
				return Nil(), fmt.Errorf("map: %w", err)
			}
			out[i] = v
		}
		return Object(out), nil
	}
	for i, item := range list {
		out[i] = item.Attr(name)
	}
	return Object(out), nil
}
