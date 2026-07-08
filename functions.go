package veloz

import "fmt"

func registerFunctions(e *Engine) {
	fns := map[string]Callable{
		"range":  fnRange,
		"max":    requireArg("max", fnMax),
		"min":    requireArg("min", fnMin),
		"length": requireArg("length", filterLength),
	}
	for name, fn := range fns {
		e.functions[name] = fn
	}
}

const maxRangeItems = 1_000_000

func fnRange(args []Value) (Value, error) {
	if len(args) < 2 {
		return Nil(), fmt.Errorf("range: expects at least start and end")
	}
	start, ok1 := asInt(args[0])
	end, ok2 := asInt(args[1])
	if !ok1 || !ok2 {
		return Nil(), fmt.Errorf("range: bounds must be numeric")
	}
	step := int64(1)
	if len(args) >= 3 {
		if s, ok := asInt(args[2]); ok && s != 0 {
			step = s
		}
	}
	var out []Value
	appendItem := func(i int64) error {
		if len(out) >= maxRangeItems {
			return fmt.Errorf("range: too many elements (limit %d)", maxRangeItems)
		}
		out = append(out, Int(i))
		return nil
	}
	if step > 0 {
		for i := start; i <= end; i += step {
			if err := appendItem(i); err != nil {
				return Nil(), err
			}
			if i > end-step {
				break
			}
		}
	} else {
		for i := start; i >= end; i += step {
			if err := appendItem(i); err != nil {
				return Nil(), err
			}
			if i < end-step {
				break
			}
		}
	}
	return Object(out), nil
}

func fnMax(args []Value) (Value, error) { return extreme(args, true) }

func fnMin(args []Value) (Value, error) { return extreme(args, false) }

func extreme(args []Value, wantMax bool) (Value, error) {
	list := args
	if len(args) == 1 {
		if l := toList(args[0]); l != nil {
			list = l
		}
	}
	if len(list) == 0 {
		return Nil(), nil
	}
	best := list[0]
	for _, v := range list[1:] {
		cmp := compareValues(v, best)
		if wantMax && cmp > 0 {
			best = v
		}
		if !wantMax && cmp < 0 {
			best = v
		}
	}
	return best, nil
}

func compareValues(a, b Value) int {
	fa, oka := asFloat(a)
	fb, okb := asFloat(b)
	if oka && okb {
		switch {
		case fa < fb:
			return -1
		case fa > fb:
			return 1
		default:
			return 0
		}
	}
	sa, sb := a.String(), b.String()
	switch {
	case sa < sb:
		return -1
	case sa > sb:
		return 1
	default:
		return 0
	}
}
