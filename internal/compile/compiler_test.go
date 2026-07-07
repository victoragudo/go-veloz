package compile

import (
	"reflect"
	"testing"

	"github.com/victoragudo/go-veloz/internal/runtime"
)

type fakeResolver struct{}

func (fakeResolver) ResolveCallable(name string, filter bool) (runtime.Callable, bool) {
	return func(args []runtime.Value) (runtime.Value, error) {
		return runtime.Nil(), nil
	}, true
}

func compileSrc(t *testing.T, src string) *runtime.Program {
	t.Helper()
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	prog, err := Compile(ast, fakeResolver{}, false)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return prog
}

func opsOf(p *runtime.Program) []runtime.Op {
	out := make([]runtime.Op, len(p.Instrs))
	for i, in := range p.Instrs {
		out[i] = in.Op
	}
	return out
}

func TestCompileArithmeticPrecedence(t *testing.T) {
	prog := compileSrc(t, `{{ 2 + 3 * 4 }}`)
	want := []runtime.Op{
		runtime.OpConst, runtime.OpConst, runtime.OpConst,
		runtime.OpMul, runtime.OpAdd, runtime.OpEcho,
	}
	if got := opsOf(prog); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompilePrint(t *testing.T) {
	prog := compileSrc(t, `{{ name }}`)
	want := []runtime.Op{runtime.OpLoadVar, runtime.OpEcho}
	if got := opsOf(prog); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompileForEmitsIterOps(t *testing.T) {
	prog := compileSrc(t, `{% for x in xs %}{{ x }}{% endfor %}`)
	ops := opsOf(prog)
	if !containsOp(ops, runtime.OpIterInit) || !containsOp(ops, runtime.OpIterNext) {
		t.Errorf("for loop missing iter ops: %v", ops)
	}
	if len(prog.Loops) != 1 {
		t.Errorf("expected 1 loop spec, got %d", len(prog.Loops))
	}
}

func TestCompileIfEmitsJumps(t *testing.T) {
	prog := compileSrc(t, `{% if a %}x{% else %}y{% endif %}`)
	ops := opsOf(prog)
	if !containsOp(ops, runtime.OpJumpFalse) || !containsOp(ops, runtime.OpJump) {
		t.Errorf("if missing jump ops: %v", ops)
	}
}

func TestCompileLocalSlots(t *testing.T) {
	prog := compileSrc(t, `{% set a = 1 %}{% set b = 2 %}{{ a }}{{ b }}`)
	if prog.NumLocals != 2 {
		t.Errorf("expected 2 local slots, got %d", prog.NumLocals)
	}
}

func TestCompileBlockRegistered(t *testing.T) {
	prog := compileSrc(t, `{% block content %}hi{% endblock %}`)
	if _, ok := prog.Blocks["content"]; !ok {
		t.Errorf("block %q not registered", "content")
	}
	if !containsOp(opsOf(prog), runtime.OpRenderBlock) {
		t.Errorf("missing RENDER_BLOCK op")
	}
}

func containsOp(ops []runtime.Op, target runtime.Op) bool {
	for _, op := range ops {
		if op == target {
			return true
		}
	}
	return false
}
