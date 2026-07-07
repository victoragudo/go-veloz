package veloz_test

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/victoragudo/go-veloz"
)

var updateGolden = flag.Bool("update", false, "rewrite golden files with the current output")

const (
	templatesDir = "testdata/templates"
	goldenDir    = "testdata/golden"
	goldenExt    = ".golden"
	tplExt       = ".tpl"
)

var sharedTemplates = []string{"layout", "footer"}

var suiteTemplates = []string{
	"expressions",
	"filters",
	"functions",
	"control",
	"loops",
	"objects",
	"escaping",
	"whitespace",
	"page",
}

type tplProduct struct {
	Name  string
	Price float64
	Tags  []string
}

func (p tplProduct) Label() string {
	return p.Name + " (" + strconv.FormatFloat(p.Price, 'f', 2, 64) + " EUR)"
}

func (p tplProduct) WithDiscount(pct float64) float64 {
	return p.Price * (1 - pct/100)
}

func suiteData() map[string]any {
	return map[string]any{
		"age":          34,
		"nickname":     "",
		"names":        []string{"Mireia", "Aleix", "Nuria"},
		"empty":        "",
		"empty_list":   []int{},
		"stock":        map[string]int{"keyboard": 12, "mouse": 30},
		"prices":       []float64{19.9, 45, 5.5},
		"temperature":  22,
		"user":         map[string]any{"active": true, "age": 41},
		"html_snippet": `<b>5 > 3 & "quotes"</b>`,
		"notes":        "first line\nsecond line",
		"grid":         [][]int{{1, 2}, {3, 4}},
		"store":        "Tienda Central",
		"product": tplProduct{
			Name:  "Mechanical keyboard",
			Price: 40,
			Tags:  []string{"peripherals", "office"},
		},
		"products": []tplProduct{
			{Name: "Mechanical keyboard", Price: 40},
			{Name: "Vertical mouse", Price: 25.5},
		},
		"order": map[string]any{
			"customer": map[string]any{"email": "nuria@tienda.es"},
			"lines":    []map[string]any{{"qty": 3}},
		},
	}
}

func readTemplateFile(tb testing.TB, name string) string {
	tb.Helper()
	src, err := os.ReadFile(filepath.Join(templatesDir, name+tplExt))
	if err != nil {
		tb.Fatal(err)
	}
	return string(src)
}

func newSuiteEngine(tb testing.TB) *veloz.Engine {
	tb.Helper()
	e := veloz.New()
	for _, name := range sharedTemplates {
		if _, err := e.Compile(name, readTemplateFile(tb, name)); err != nil {
			tb.Fatalf("compile %s: %v", name, err)
		}
	}
	return e
}

func TestTemplateSuite(t *testing.T) {
	e := newSuiteEngine(t)
	data := suiteData()

	for _, name := range suiteTemplates {
		t.Run(name, func(t *testing.T) {
			tmpl, err := e.Compile(name, readTemplateFile(t, name))
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			got, err := tmpl.Render(data)
			if err != nil {
				t.Fatalf("render: %v", err)
			}

			goldenPath := filepath.Join(goldenDir, name+goldenExt)
			if *updateGolden {
				if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("missing golden file, run go test -run TestTemplateSuite -update: %v", err)
			}
			if got != string(want) {
				t.Errorf("output does not match golden file\n%s", diffLines(string(want), got))
			}
		})
	}
}

func diffLines(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	var b strings.Builder
	for i := 0; i < len(wantLines) || i < len(gotLines); i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w != g {
			b.WriteString("line " + strconv.Itoa(i+1) + "\n")
			b.WriteString("  want: " + strconv.Quote(w) + "\n")
			b.WriteString("  got:  " + strconv.Quote(g) + "\n")
		}
	}
	return b.String()
}
