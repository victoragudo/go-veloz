package veloz_test

import (
	"bytes"
	htmltemplate "html/template"
	"testing"
	texttemplate "text/template"

	"github.com/victoragudo/go-veloz"
)

type benchUser struct {
	Name  string
	Email string
	Admin bool
}

func benchUsers(n int) []benchUser {
	names := []string{"Mireia", "Aleix", "Nuria", "Oriol", "Laia", "Marc", "Julia", "Pau"}
	users := make([]benchUser, n)
	for i := range users {
		who := names[i%len(names)]
		users[i] = benchUser{
			Name:  who,
			Email: who + "@example.com",
			Admin: i%3 == 0,
		}
	}
	return users
}

const velozSrc = `<ul>{% for u in users %}<li>{{ u.Name }} - {{ u.Email }}{% if u.Admin %} (admin){% endif %}</li>{% endfor %}</ul>`

const stdSrc = `<ul>{{range .Users}}<li>{{.Name}} - {{.Email}}{{if .Admin}} (admin){{end}}</li>{{end}}</ul>`

func BenchmarkVeloz(b *testing.B) {
	e := veloz.New(veloz.WithAutoescape(false))
	tmpl := e.MustCompile("bench", velozSrc)
	data := map[string]any{"users": benchUsers(50)}
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.RenderTo(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTextTemplate(b *testing.B) {
	tmpl, err := texttemplate.New("bench").Parse(stdSrc)
	if err != nil {
		b.Fatal(err)
	}
	data := struct{ Users []benchUser }{Users: benchUsers(50)}
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVelozAutoescape(b *testing.B) {
	e := veloz.New()
	tmpl := e.MustCompile("bench", velozSrc)
	data := map[string]any{"users": benchUsers(50)}
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.RenderTo(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTMLTemplate(b *testing.B) {
	tmpl, err := htmltemplate.New("bench").Parse(stdSrc)
	if err != nil {
		b.Fatal(err)
	}
	data := struct{ Users []benchUser }{Users: benchUsers(50)}
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatal(err)
		}
	}
}
