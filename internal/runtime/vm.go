package runtime

import (
	"fmt"
	"reflect"
	"sort"
)

type LoopState struct {
	i      int
	length int
}

func (l *LoopState) VelozAttr(name string) (Value, bool) {
	switch name {
	case "index":
		return Int(int64(l.i + 1)), true
	case "index0":
		return Int(int64(l.i)), true
	case "revindex":
		return Int(int64(l.length - l.i)), true
	case "revindex0":
		return Int(int64(l.length - l.i - 1)), true
	case "first":
		return Bool(l.i == 0), true
	case "last":
		return Bool(l.i == l.length-1), true
	case "length":
		return Int(int64(l.length)), true
	}
	return Nil(), false
}

type iterState struct {
	spec   *LoopSpec
	items  reflect.Value
	keys   []reflect.Value
	isMap  bool
	i      int
	length int
}

type frame struct {
	stack  []Value
	locals []Value
	iters  []iterState
}

type Interp struct {
	out    []byte
	loader Loader
	blocks map[string]*Program
	frames []frame
	depth  int
}

func NewInterp(loader Loader) *Interp {
	return &Interp{loader: loader, out: make([]byte, 0, 512)}
}

func (ip *Interp) Run(prog *Program, data Value, blocks map[string]*Program) ([]byte, error) {
	ip.out = ip.out[:0]
	ip.blocks = blocks
	ip.depth = 0
	if err := ip.exec(prog, data); err != nil {
		return nil, err
	}
	return ip.out, nil
}

const (
	minStackCap   = 32
	maxFrameDepth = 64
)

func (ip *Interp) exec(prog *Program, data Value) error {
	if ip.depth >= maxFrameDepth {
		return fmt.Errorf("template nesting deeper than %d frames, check for include or inheritance cycles", maxFrameDepth)
	}
	if ip.depth >= len(ip.frames) {
		ip.frames = append(ip.frames, frame{})
	}
	idx := ip.depth
	ip.depth++
	err := ip.runFrame(prog, data, idx)
	ip.depth--
	return err
}

func (ip *Interp) frameLocals(idx, n int) []Value {
	fr := &ip.frames[idx]
	if cap(fr.locals) < n {
		fr.locals = make([]Value, n)
		return fr.locals
	}
	locals := fr.locals[:n]
	for i := range locals {
		locals[i] = Value{}
	}
	return locals
}

func (ip *Interp) frameStack(idx int) []Value {
	fr := &ip.frames[idx]
	if cap(fr.stack) < minStackCap {
		fr.stack = make([]Value, 0, minStackCap)
	}
	return fr.stack[:0]
}

func buildIter(spec *LoopSpec, coll Value) iterState {
	it := iterState{spec: spec}
	if coll.kind != KindObject {
		return it
	}
	rv := reflect.ValueOf(coll.obj)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return it
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		it.items = rv
		it.length = rv.Len()
	case reflect.Map:
		it.isMap = true
		it.items = rv
		keys := rv.MapKeys()
		sort.Slice(keys, func(a, b int) bool {
			return fmt.Sprint(keys[a].Interface()) < fmt.Sprint(keys[b].Interface())
		})
		it.keys = keys
		it.length = len(keys)
	default:
		return it
	}
	return it
}

func (ip *Interp) runFrame(prog *Program, data Value, frameIdx int) error {
	instrs := prog.Instrs
	consts := prog.Consts
	callables := prog.Callables
	loops := prog.Loops
	autoescape := prog.Autoescape

	locals := ip.frameLocals(frameIdx, prog.NumLocals)
	stack := ip.frameStack(frameIdx)
	iters := ip.frames[frameIdx].iters[:0]

	out := ip.out
	pc := 0

	for pc < len(instrs) {
		in := instrs[pc]
		switch in.Op {
		case OpText:
			out = append(out, consts[in.Arg].str...)

		case OpConst:
			stack = append(stack, consts[in.Arg])

		case OpNil:
			stack = append(stack, Nil())

		case OpTrue:
			stack = append(stack, Bool(true))

		case OpFalse:
			stack = append(stack, Bool(false))

		case OpLoadVar:
			stack = append(stack, resolveAttr(data, consts[in.Arg].str))

		case OpLoadLocal:
			stack = append(stack, locals[in.Arg])

		case OpStoreLocal:
			locals[in.Arg] = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

		case OpAttr:
			top := len(stack) - 1
			stack[top] = resolveAttr(stack[top], consts[in.Arg].str)

		case OpIndex:
			idx := stack[len(stack)-1]
			obj := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = resolveIndex(obj, idx)

		case OpCall:
			idx, argc := unpackCall(in.Arg)
			base := len(stack) - argc
			res, err := callables[idx](stack[base:])
			if err != nil {
				ip.out = out
				return err
			}
			stack = stack[:base]
			stack = append(stack, res)

		case OpCallMethod:
			nameIdx, argc := unpackCall(in.Arg)
			base := len(stack) - argc
			recv := stack[base-1]
			res, err := callMethod(recv, consts[nameIdx].str, stack[base:])
			if err != nil {
				ip.out = out
				return err
			}
			stack = stack[:base-1]
			stack = append(stack, res)

		case OpMakeList:
			base := len(stack) - in.Arg
			elems := make([]Value, in.Arg)
			copy(elems, stack[base:])
			stack = stack[:base]
			stack = append(stack, Object(elems))

		case OpMakeMap:
			base := len(stack) - in.Arg*2
			m := make(map[string]Value, in.Arg)
			for i := 0; i < in.Arg; i++ {
				m[stack[base+2*i].String()] = stack[base+2*i+1]
			}
			stack = stack[:base]
			stack = append(stack, Object(m))

		case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpPow:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			res, err := binaryArith(in.Op, a, b)
			if err != nil {
				ip.out = out
				return err
			}
			stack[len(stack)-1] = res

		case OpNeg:
			res, err := arithNeg(stack[len(stack)-1])
			if err != nil {
				ip.out = out
				return err
			}
			stack[len(stack)-1] = res

		case OpConcat:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = Str(a.String() + b.String())

		case OpEq:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = Bool(valueEq(a, b))

		case OpNeq:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = Bool(!valueEq(a, b))

		case OpLt, OpGt, OpLte, OpGte:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			cmp, err := valueCmp(a, b)
			if err != nil {
				ip.out = out
				return err
			}
			res, err := cmpResult(in.Op, cmp)
			if err != nil {
				ip.out = out
				return err
			}
			stack[len(stack)-1] = Bool(res)

		case OpIn:
			seq := stack[len(stack)-1]
			item := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = Bool(contains(item, seq))

		case OpNotIn:
			seq := stack[len(stack)-1]
			item := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = Bool(!contains(item, seq))

		case OpNot:
			stack[len(stack)-1] = Bool(!stack[len(stack)-1].IsTruthy())

		case OpEcho:
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if raw, ok := isSafe(v); ok {
				out = append(out, raw...)
			} else if autoescape {
				out = writeEscaped(out, v.String())
			} else {
				out = append(out, v.String()...)
			}

		case OpPop:
			stack = stack[:len(stack)-1]

		case OpJump:
			pc = in.Arg
			continue

		case OpJumpFalse:
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !v.IsTruthy() {
				pc = in.Arg
				continue
			}

		case OpJumpIfFalseOrPop:
			if !stack[len(stack)-1].IsTruthy() {
				pc = in.Arg
				continue
			}
			stack = stack[:len(stack)-1]

		case OpJumpIfTrueOrPop:
			if stack[len(stack)-1].IsTruthy() {
				pc = in.Arg
				continue
			}
			stack = stack[:len(stack)-1]

		case OpIterInit:
			spec := &loops[in.Arg]
			coll := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			it := buildIter(spec, coll)
			if it.length == 0 {
				pc = spec.ElseTarget
				continue
			}
			iters = append(iters, it)

		case OpIterNext:
			it := &iters[len(iters)-1]
			if it.i >= it.length {
				target := it.spec.EndTarget
				iters = iters[:len(iters)-1]
				pc = target
				continue
			}
			if it.isMap {
				k := it.keys[it.i]
				locals[it.spec.ValSlot] = FromAny(it.items.MapIndex(k).Interface())
				if it.spec.KeySlot >= 0 {
					locals[it.spec.KeySlot] = FromAny(k.Interface())
				}
			} else {
				locals[it.spec.ValSlot] = FromAny(it.items.Index(it.i).Interface())
				if it.spec.KeySlot >= 0 {
					locals[it.spec.KeySlot] = Int(int64(it.i))
				}
			}
			if it.spec.LoopSlot >= 0 {
				locals[it.spec.LoopSlot] = Object(&LoopState{i: it.i, length: it.length})
			}
			it.i++

		case OpRenderBlock:
			name := consts[in.Arg].str
			block, ok := ip.blocks[name]
			if !ok {
				break
			}
			ip.out = out
			if err := ip.exec(block, data); err != nil {
				return err
			}
			out = ip.out

		case OpInclude:
			name := consts[in.Arg].str
			if ip.loader == nil {
				ip.out = out
				return fmt.Errorf("include %q: no loader configured", name)
			}
			sub, err := ip.loader.LoadProgram(name)
			if err != nil {
				ip.out = out
				return fmt.Errorf("include %q: %w", name, err)
			}
			ip.out = out
			if err := ip.exec(sub, data); err != nil {
				return err
			}
			out = ip.out

		default:
			ip.out = out
			return fmt.Errorf("unknown opcode %d", in.Op)
		}
		pc++
	}

	fr := &ip.frames[frameIdx]
	fr.stack = stack[:0]
	fr.iters = iters[:0]
	ip.out = out
	return nil
}

func binaryArith(op Op, a, b Value) (Value, error) {
	switch op {
	case OpAdd:
		return arithAdd(a, b)
	case OpSub:
		return arithSub(a, b)
	case OpMul:
		return arithMul(a, b)
	case OpDiv:
		return arithDiv(a, b)
	case OpMod:
		return arithMod(a, b)
	case OpPow:
		return arithPow(a, b)
	default:
		return Nil(), fmt.Errorf("invalid arithmetic op")
	}
}

func cmpResult(op Op, cmp int) (bool, error) {
	switch op {
	case OpLt:
		return cmp < 0, nil
	case OpGt:
		return cmp > 0, nil
	case OpLte:
		return cmp <= 0, nil
	case OpGte:
		return cmp >= 0, nil
	default:
		return false, fmt.Errorf("invalid comparison op %d", op)
	}
}
