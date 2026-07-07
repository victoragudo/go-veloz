package veloz_test

import (
	"bytes"
	"fmt"
	"testing"
	texttemplate "text/template"

	"veloz"
)

var scaleSizes = []int{1, 100, 1000, 10000}

func BenchmarkScaleVeloz(b *testing.B) {
	e := veloz.New(veloz.WithAutoescape(false))
	tmpl := e.MustCompile("scale", velozSrc)
	for _, n := range scaleSizes {
		data := map[string]any{"users": benchUsers(n)}
		b.Run(fmt.Sprintf("users_%d", n), func(b *testing.B) {
			var buf bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := tmpl.RenderTo(&buf, data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkScaleTextTemplate(b *testing.B) {
	tmpl := texttemplate.Must(texttemplate.New("scale").Parse(stdSrc))
	for _, n := range scaleSizes {
		data := struct{ Users []benchUser }{Users: benchUsers(n)}
		b.Run(fmt.Sprintf("users_%d", n), func(b *testing.B) {
			var buf bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := tmpl.Execute(&buf, data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCompileVeloz(b *testing.B) {
	e := veloz.New(veloz.WithAutoescape(false))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := e.Compile("scale", velozSrc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileTextTemplate(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := texttemplate.New("scale").Parse(stdSrc); err != nil {
			b.Fatal(err)
		}
	}
}
