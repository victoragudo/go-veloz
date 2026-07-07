package veloz

import (
	"io"

	"github.com/victoragudo/go-veloz/internal/runtime"
)

type Template struct {
	eng  *Engine
	prog *runtime.Program
}

func (t *Template) Name() string { return t.prog.Name }

func (t *Template) Render(data any) (string, error) {
	b, err := t.RenderBytes(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *Template) RenderBytes(data any) ([]byte, error) {
	root, blocks, err := t.eng.resolve(t.prog)
	if err != nil {
		return nil, err
	}
	ip := t.eng.getInterp()
	out, err := ip.Run(root, runtime.FromAny(data), blocks)
	if err != nil {
		t.eng.putInterp(ip)
		return nil, err
	}
	res := make([]byte, len(out))
	copy(res, out)
	t.eng.putInterp(ip)
	return res, nil
}

func (t *Template) MustRender(data any) string {
	s, err := t.Render(data)
	if err != nil {
		panic(err)
	}
	return s
}

func (t *Template) RenderTo(w io.Writer, data any) error {
	root, blocks, err := t.eng.resolve(t.prog)
	if err != nil {
		return err
	}
	ip := t.eng.getInterp()
	out, err := ip.Run(root, runtime.FromAny(data), blocks)
	if err != nil {
		t.eng.putInterp(ip)
		return err
	}
	_, werr := w.Write(out)
	t.eng.putInterp(ip)
	return werr
}
