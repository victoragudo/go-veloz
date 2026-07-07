package veloz_test

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	texttemplate "text/template"

	"veloz"
)

const goTemplatesDir = "testdata/gotemplates"

var suiteScaleSizes = []int{1, 10, 100, 1000, 10000, 100000}

type goSuiteData struct {
	Age         int
	Nickname    string
	Missing     string
	Names       []string
	Empty       string
	EmptyList   []int
	Stock       map[string]int
	Prices      []float64
	Temperature int
	User        map[string]any
	Snippet     string
	Notes       string
	Grid        [][]int
	Store       string
	Title       string
	Product     tplProduct
	Products    []tplProduct
	Order       map[string]any
}

func goSuiteDataValue() goSuiteData {
	return goSuiteData{
		Age:         34,
		Nickname:    "",
		Missing:     "",
		Names:       []string{"Mireia", "Aleix", "Nuria"},
		Empty:       "",
		EmptyList:   []int{},
		Stock:       map[string]int{"keyboard": 12, "mouse": 30},
		Prices:      []float64{19.9, 45, 5.5},
		Temperature: 22,
		User:        map[string]any{"active": true, "age": 41},
		Snippet:     `<b>5 > 3 & "quotes"</b>`,
		Notes:       "first line\nsecond line",
		Grid:        [][]int{{1, 2}, {3, 4}},
		Store:       "Tienda Central",
		Product: tplProduct{
			Name:  "Mechanical keyboard",
			Price: 40,
			Tags:  []string{"peripherals", "office"},
		},
		Products: []tplProduct{
			{Name: "Mechanical keyboard", Price: 40},
			{Name: "Vertical mouse", Price: 25.5},
		},
		Order: map[string]any{
			"customer": map[string]any{"email": "nuria@tienda.es"},
			"lines":    []map[string]any{{"qty": 3}},
		},
	}
}

func goSuiteFuncs() texttemplate.FuncMap {
	fm := goFeatFuncs()
	fm["divf"] = func(a, b float64) float64 { return a / b }
	fm["modf"] = math.Mod
	fm["eqbi"] = func(b bool, i int) bool {
		if b {
			return i == 1
		}
		return i == 0
	}
	fm["instr"] = func(sub, s string) bool { return strings.Contains(s, sub) }
	fm["trimc"] = strings.Trim
	fm["revstr"] = func(s string) string {
		r := []rune(s)
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r)
	}
	fm["keysint"] = func(m map[string]int) []string {
		out := make([]string, 0, len(m))
		for k := range m {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}
	fm["seqstep"] = func(start, end, step int) []int {
		var out []int
		if step > 0 {
			for i := start; i <= end; i += step {
				out = append(out, i)
			}
		} else if step < 0 {
			for i := start; i >= end; i += step {
				out = append(out, i)
			}
		}
		return out
	}
	fm["maxfs"] = func(xs []float64) float64 {
		best := xs[0]
		for _, x := range xs[1:] {
			if x > best {
				best = x
			}
		}
		return best
	}
	return fm
}

func readGoTemplateFile(tb testing.TB, name string) string {
	tb.Helper()
	src, err := os.ReadFile(filepath.Join(goTemplatesDir, name+tplExt))
	if err != nil {
		tb.Fatal(err)
	}
	return string(src)
}

func compileSuiteVeloz(tb testing.TB) map[string]*veloz.Template {
	tb.Helper()
	e := newSuiteEngine(tb)
	out := make(map[string]*veloz.Template, len(suiteTemplates))
	for _, name := range suiteTemplates {
		tmpl, err := e.Compile(name, readTemplateFile(tb, name))
		if err != nil {
			tb.Fatalf("compile veloz %s: %v", name, err)
		}
		out[name] = tmpl
	}
	return out
}

func compileSuiteGo(tb testing.TB) map[string]*texttemplate.Template {
	tb.Helper()
	funcs := goSuiteFuncs()
	out := make(map[string]*texttemplate.Template, len(suiteTemplates))
	for _, name := range suiteTemplates {
		tmpl, err := texttemplate.New(name).Funcs(funcs).Parse(readGoTemplateFile(tb, name))
		if err != nil {
			tb.Fatalf("compile go %s: %v", name, err)
		}
		out[name] = tmpl
	}
	return out
}

func TestGoTemplateSuiteParity(t *testing.T) {
	templates := compileSuiteGo(t)
	data := goSuiteDataValue()

	for _, name := range suiteTemplates {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := templates[name].Execute(&buf, data); err != nil {
				t.Fatalf("render: %v", err)
			}
			want, err := os.ReadFile(filepath.Join(goldenDir, name+goldenExt))
			if err != nil {
				t.Fatal(err)
			}
			if buf.String() != string(want) {
				t.Errorf("go template output does not match golden file\n%s", diffLines(string(want), buf.String()))
			}
		})
	}
}

func BenchmarkSuiteVeloz(b *testing.B) {
	templates := compileSuiteVeloz(b)
	data := suiteData()
	for _, name := range suiteTemplates {
		b.Run(name, func(b *testing.B) {
			tmpl := templates[name]
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

func BenchmarkSuiteTextTemplate(b *testing.B) {
	templates := compileSuiteGo(b)
	data := goSuiteDataValue()
	for _, name := range suiteTemplates {
		b.Run(name, func(b *testing.B) {
			tmpl := templates[name]
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

func scaledNames(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("client-%d", i+1)
	}
	return out
}

func scaledProducts(n int) []tplProduct {
	out := make([]tplProduct, n)
	for i := range out {
		out[i] = tplProduct{
			Name:  fmt.Sprintf("Product %d", i+1),
			Price: float64(i%90) + 9.5,
		}
	}
	return out
}

func BenchmarkSuiteScaleVeloz(b *testing.B) {
	templates := compileSuiteVeloz(b)
	for _, n := range suiteScaleSizes {
		data := suiteData()
		data["names"] = scaledNames(n)
		data["products"] = scaledProducts(n)
		for _, name := range []string{"loops", "page"} {
			b.Run(fmt.Sprintf("%s_%d", name, n), func(b *testing.B) {
				tmpl := templates[name]
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
}

func BenchmarkSuiteScaleTextTemplate(b *testing.B) {
	templates := compileSuiteGo(b)
	for _, n := range suiteScaleSizes {
		data := goSuiteDataValue()
		data.Names = scaledNames(n)
		data.Products = scaledProducts(n)
		for _, name := range []string{"loops", "page"} {
			b.Run(fmt.Sprintf("%s_%d", name, n), func(b *testing.B) {
				tmpl := templates[name]
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
}
