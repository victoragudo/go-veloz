package veloz

import (
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/victoragudo/go-veloz/internal/compile"
	"github.com/victoragudo/go-veloz/internal/runtime"
)

type templateEntry struct {
	prog    *runtime.Program
	modTime time.Time
	fromFS  bool
}

type Engine struct {
	mu         sync.RWMutex
	filters    map[string]Callable
	functions  map[string]Callable
	templates  map[string]*templateEntry
	fsys       fs.FS
	reload     bool
	autoescape bool
	pool       sync.Pool
}

type Option func(*Engine)

func WithAutoescape(on bool) Option {
	return func(e *Engine) { e.autoescape = on }
}

func WithFS(fsys fs.FS) Option {
	return func(e *Engine) { e.fsys = fsys }
}

func WithReload(on bool) Option {
	return func(e *Engine) { e.reload = on }
}

func New(opts ...Option) *Engine {
	e := &Engine{
		filters:    map[string]Callable{},
		functions:  map[string]Callable{},
		templates:  map[string]*templateEntry{},
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
	first, second := e.functions, e.filters
	if filter {
		first, second = e.filters, e.functions
	}
	if fn, ok := first[name]; ok {
		return fn, true
	}
	if fn, ok := second[name]; ok {
		return fn, true
	}
	return nil, false
}

func (e *Engine) Compile(name, src string) (*Template, error) {
	prog, err := e.compileSource(name, src, time.Time{}, false)
	if err != nil {
		return nil, err
	}
	return &Template{eng: e, prog: prog}, nil
}

func (e *Engine) MustCompile(name, src string) *Template {
	t, err := e.Compile(name, src)
	if err != nil {
		panic(err)
	}
	return t
}

func (e *Engine) Load(name string) (*Template, error) {
	prog, err := e.LoadProgram(name)
	if err != nil {
		return nil, err
	}
	return &Template{eng: e, prog: prog}, nil
}

func (e *Engine) MustLoad(name string) *Template {
	t, err := e.Load(name)
	if err != nil {
		panic(err)
	}
	return t
}

func (e *Engine) Get(name string) (*Template, bool) {
	e.mu.RLock()
	ent, ok := e.templates[name]
	e.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return &Template{eng: e, prog: ent.prog}, true
}

func (e *Engine) LoadProgram(name string) (*runtime.Program, error) {
	e.mu.RLock()
	ent, ok := e.templates[name]
	e.mu.RUnlock()
	if ok && (!e.reload || !ent.fromFS) {
		return ent.prog, nil
	}
	if e.fsys == nil {
		if ok {
			return ent.prog, nil
		}
		return nil, fmt.Errorf("template %q not found and no filesystem configured", name)
	}
	if ok && e.reload {
		info, err := fs.Stat(e.fsys, name)
		if err == nil && !info.ModTime().After(ent.modTime) {
			return ent.prog, nil
		}
	}
	return e.loadFromFS(name)
}

func (e *Engine) loadFromFS(name string) (*runtime.Program, error) {
	src, err := fs.ReadFile(e.fsys, name)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", name, err)
	}
	modTime := time.Time{}
	if info, statErr := fs.Stat(e.fsys, name); statErr == nil {
		modTime = info.ModTime()
	}
	return e.compileSource(name, string(src), modTime, true)
}

func (e *Engine) compileSource(name, src string, modTime time.Time, fromFS bool) (*runtime.Program, error) {
	ast, err := compile.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", name, err)
	}
	prog, err := compile.Compile(name, ast, e, e.autoescape)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", name, err)
	}
	e.mu.Lock()
	e.templates[name] = &templateEntry{prog: prog, modTime: modTime, fromFS: fromFS}
	e.mu.Unlock()
	return prog, nil
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
		parent, err := e.LoadProgram(cur.Parent)
		if err != nil {
			return nil, nil, fmt.Errorf("parent of %q: %w", cur.Name, err)
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
