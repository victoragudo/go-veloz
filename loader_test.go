package veloz_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/victoragudo/go-veloz"
)

func suiteFS() fstest.MapFS {
	return fstest.MapFS{
		"layout.tpl":          {Data: []byte(`[{% block body %}{% endblock %}] {% include "partials/footer.tpl" %}`)},
		"partials/footer.tpl": {Data: []byte(`footer of {{ shop }}`)},
		"pages/home.tpl":      {Data: []byte(`{% extends "../layout.tpl" %}{% block body %}home of {{ shop }}{% endblock %}`)},
		"pages/about.tpl":     {Data: []byte(`{% include "./team.tpl" %} and {% include "../partials/footer.tpl" %}`)},
		"pages/team.tpl":      {Data: []byte(`team page`)},
		"broken/oops.tpl":     {Data: []byte(`{{ name | nonexistentfilter }}`)},
		"broken/includer.tpl": {Data: []byte(`{% include "./oops.tpl" %}`)},
		"cycle/self.tpl":      {Data: []byte(`{% include "./self.tpl" %}`)},
	}
}

func TestLoadFromFS(t *testing.T) {
	e := veloz.New(veloz.WithFS(suiteFS()))
	tmpl, err := e.Load("pages/home.tpl")
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(map[string]any{"shop": "Tienda Central"})
	if err != nil {
		t.Fatal(err)
	}
	want := "[home of Tienda Central] footer of Tienda Central"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRelativeIncludes(t *testing.T) {
	e := veloz.New(veloz.WithFS(suiteFS()))
	tmpl, err := e.Load("pages/about.tpl")
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(map[string]any{"shop": "Tienda Central"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "team page and footer of Tienda Central" {
		t.Errorf("got %q", out)
	}
}

func TestLoadUnknownTemplate(t *testing.T) {
	e := veloz.New(veloz.WithFS(suiteFS()))
	if _, err := e.Load("pages/missing.tpl"); err == nil || !strings.Contains(err.Error(), "pages/missing.tpl") {
		t.Errorf("expected error naming the template, got %v", err)
	}

	plain := veloz.New()
	if _, err := plain.Load("anything.tpl"); err == nil || !strings.Contains(err.Error(), "no filesystem") {
		t.Errorf("expected no-filesystem error, got %v", err)
	}
}

func TestIncludeCompileErrorPropagates(t *testing.T) {
	e := veloz.New(veloz.WithFS(suiteFS()))
	tmpl, err := e.Load("broken/includer.tpl")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmpl.Render(nil)
	if err == nil || !strings.Contains(err.Error(), "nonexistentfilter") {
		t.Errorf("expected the include to surface the compile error, got %v", err)
	}
}

func TestIncludeCycleFails(t *testing.T) {
	e := veloz.New(veloz.WithFS(suiteFS()))
	tmpl, err := e.Load("cycle/self.tpl")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmpl.Render(nil)
	if err == nil || !strings.Contains(err.Error(), "nesting deeper") {
		t.Errorf("expected a nesting depth error, got %v", err)
	}
}

func TestReloadPicksUpChanges(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "greeting.tpl")
	if err := os.WriteFile(file, []byte("version one"), 0o644); err != nil {
		t.Fatal(err)
	}

	e := veloz.New(veloz.WithFS(os.DirFS(dir)), veloz.WithReload(true))
	tmpl, err := e.Load("greeting.tpl")
	if err != nil {
		t.Fatal(err)
	}
	if out, _ := tmpl.Render(nil); out != "version one" {
		t.Fatalf("got %q", out)
	}

	if err := os.WriteFile(file, []byte("version two"), 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(file, future, future); err != nil {
		t.Fatal(err)
	}

	tmpl2, err := e.Load("greeting.tpl")
	if err != nil {
		t.Fatal(err)
	}
	if out, _ := tmpl2.Render(nil); out != "version two" {
		t.Errorf("reload did not pick up the new content, got %q", out)
	}
}

func TestNoReloadKeepsCache(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "greeting.tpl")
	if err := os.WriteFile(file, []byte("version one"), 0o644); err != nil {
		t.Fatal(err)
	}

	e := veloz.New(veloz.WithFS(os.DirFS(dir)))
	if _, err := e.Load("greeting.tpl"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("version two"), 0o644); err != nil {
		t.Fatal(err)
	}
	tmpl, err := e.Load("greeting.tpl")
	if err != nil {
		t.Fatal(err)
	}
	if out, _ := tmpl.Render(nil); out != "version one" {
		t.Errorf("cache was not used, got %q", out)
	}
}

func TestErrorPositions(t *testing.T) {
	e := veloz.New()
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"unknown_filter", "line one\n  {{ name | slugg }}", "bad.tpl:2:13: unknown filter \"slugg\""},
		{"unknown_function", "{{ slugify(name) }}", "bad.tpl:1:11: unknown function \"slugify\""},
		{"parse_error", "{% if x %}\n{{ 1 + }}", "bad.tpl:2:8:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := e.Compile("bad.tpl", tc.src)
			if err == nil {
				t.Fatal("expected a compile error")
			}
			if !strings.HasPrefix(err.Error(), tc.want) {
				t.Errorf("got %q, want prefix %q", err.Error(), tc.want)
			}
		})
	}
}
