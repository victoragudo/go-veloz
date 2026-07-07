package veloz

import (
	"fmt"
	"sync"

	"github.com/victoragudo/go-veloz/internal/compile"
	"github.com/victoragudo/go-veloz/internal/runtime"
)

type Engine struct {
	mu         sync.RWMutex
	filters    map[string]Callable
	functions  map[string]Callable
	templates  map[string]*runtime.Program
	autoescape bool
	pool       sync.Pool
}

type Option func(*Engine)

func WithAutoescape(on bool) Option {
	return func(e *Engine) { e.autoescape = on }
}

func New(opts ...Option) *Engine {
	e := &Engine{
		filters:    map[string]Callable{},
		functions:  map[string]Callable{},
		templates:  map[string]*runtime.Program{},
		autoescape: true,
	}
	registerBuiltins(e)
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) RegisterFilter(name string, fn Callable) {
	e.mu.Lock()
	e.filters[name] = fn
	e.mu.Unlock()
}

func (e *Engine) RegisterFunction(name string, fn Callable) {
	e.mu.Lock()
	e.functions[name] = fn
	e.mu.Unlock()
}

func (e *Engine) ResolveCallable(name string, filter bool) (runtime.Callable, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if filter {
		if fn, ok := e.filters[name]; ok {
			return fn, true
		}
		if fn, ok := e.functions[name]; ok {
			return fn, true
		}
		return nil, false
	}
	if fn, ok := e.functions[name]; ok {
		return fn, true
	}
	if fn, ok := e.filters[name]; ok {
		return fn, true
	}
	return nil, false
}

func (e *Engine) Load(name string) (*runtime.Program, bool) {
	e.mu.RLock()
	p, ok := e.templates[name]
	e.mu.RUnlock()
	return p, ok
}

func (e *Engine) Compile(name, src string) (*Template, error) {
	ast, err := compile.Parse(src)
	if err != nil {
		return nil, err
	}
	prog, err := compile.Compile(ast, e, e.autoescape)
	if err != nil {
		return nil, err
	}
	prog.Name = name
	e.mu.Lock()
	e.templates[name] = prog
	e.mu.Unlock()
	return &Template{eng: e, prog: prog}, nil
}

func (e *Engine) MustCompile(name, src string) *Template {
	t, err := e.Compile(name, src)
	if err != nil {
		panic(err)
	}
	return t
}

func (e *Engine) Get(name string) (*Template, bool) {
	e.mu.RLock()
	p, ok := e.templates[name]
	e.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return &Template{eng: e, prog: p}, true
}

func (e *Engine) resolve(prog *runtime.Program) (*runtime.Program, map[string]*runtime.Program, error) {
	if prog.Parent == "" {
		return prog, prog.Blocks, nil
	}
	merged := map[string]*runtime.Program{}
	cur := prog
	seen := map[string]bool{}
	for {
		for name, b := range cur.Blocks {
			if _, ok := merged[name]; !ok {
				merged[name] = b
			}
		}
		if cur.Parent == "" {
			return cur, merged, nil
		}
		if seen[cur.Parent] {
			return nil, nil, fmt.Errorf("cyclic template inheritance involving %q", cur.Parent)
		}
		seen[cur.Parent] = true
		parent, ok := e.Load(cur.Parent)
		if !ok {
			return nil, nil, fmt.Errorf("parent template %q not found", cur.Parent)
		}
		cur = parent
	}
}

func (e *Engine) getInterp() *runtime.Interp {
	if v := e.pool.Get(); v != nil {
		return v.(*runtime.Interp)
	}
	return runtime.NewInterp(e)
}

func (e *Engine) putInterp(ip *runtime.Interp) {
	e.pool.Put(ip)
}
