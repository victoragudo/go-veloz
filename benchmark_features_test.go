package veloz_test

import (
	"bytes"
	"html"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
	texttemplate "text/template"
	"unicode"
	"unicode/utf8"

	"veloz"
)

type featUser struct {
	Name  string
	Email string
	Admin bool
	Score float64
	Tags  []string
	Bio   string
	Note  string
}

func (u featUser) Label() string {
	return u.Name + "(" + strconv.FormatFloat(u.Score, 'f', -1, 64) + ")"
}

func (u featUser) Discounted(pct float64) float64 {
	return u.Score * (1 - pct/100)
}

func featUsers() []featUser {
	return []featUser{
		{Name: "Mireia", Email: "MIREIA@Example.com", Admin: true, Score: 91.5, Tags: []string{"backend", "go"}, Bio: `<b>likes "go" & templates</b>`},
		{Name: "Aleix", Email: "Aleix@Example.com", Admin: false, Score: 73, Tags: []string{"frontend"}, Bio: "plain bio", Note: "on holidays"},
		{Name: "Nuria", Email: "NURIA@EXAMPLE.COM", Admin: false, Score: 44, Tags: []string{"qa", "e2e"}, Bio: "a<b", Note: ""},
	}
}

func featMeta() map[string]string {
	return map[string]string{"env": "prod", "region": "eu-west"}
}

const velozLayoutSrc = `REPORT|{% block body %}{% endblock %}|END v{{ version }}`

const velozFooterSrc = `pie-{{ version }}`

const velozFeatSrc = `{% extends "layout" %}{% block body %}
{% for u in users %}{{ loop.index }}/{{ loop.length }} {{ loop.first ? "*" : (loop.last ? "$" : "-") }} {{ u.Name | upper }} <{{ u.Email | lower }}>{% if u.Admin %} [ADMIN]{% elseif u.Score >= 60 %} [PRO]{% else %} [BASIC]{% endif %} tags={{ u.Tags | join(",") }} rev={{ u.Tags | reverse | join(",") }} n={{ u.Tags | length }} label={{ u.Label() }} disc={{ u.Discounted(10) | round(2) }} bio={{ u.Bio }} raw={{ u.Bio | raw }} note={{ u.Note | default("none") }}
{% endfor %}empty=[{% for x in nothing %}x{% else %}vacio{% endfor %}]
math={{ (2 ** 5 + 4) * 3 % 100 }} abs={{ (-7.5) | abs }} concat={{ "go" ~ "-" ~ "veloz" }} logic={{ 2 < 3 and not false or 1 > 2 }}
ternary={{ 18 >= 18 ? "adult" : "minor" }} elvis={{ "" ?: "anon" }} in={{ 2 in [1, 2, 3] }} notin={{ 9 not in [1, 2, 3] }}
idx={{ [10, 20, 30][1] }} neg={{ [10, 20, 30][-1] }} city={{ {ciudad: "Barcelona", pais: "ES"}.ciudad }}
cap={{ "hola mundo" | capitalize }} title={{ "punto de venta" | title }} trim=[{{ "  centrado  " | trim }}] replace={{ "gato" | replace("g", "p") }} split={{ "a;b;c" | split(";") | join("+") }}
first={{ "veloz" | first }} last={{ "veloz" | last }} nl2br={{ salto | nl2br }}
range={{ range(1, 5) | join("-") }} max={{ max(3, 8, 2) }} min={{ min([4, 1, 7]) }} len={{ length("catalunya") }}
keys={% for k in meta | keys %}{{ k }};{% endfor %} kv={% for k, v in meta %}{{ k }}={{ v }};{% endfor %}
{% set total = 0 %}{% for u in users %}{% set total = total + u.Score %}{% endfor %}total={{ total }}
ws=[{%- if true -%} ok {%- endif -%}]
inc=[{% include "footer" %}]{% endblock %}`

const goFeatSrc = `{{define "footer"}}pie-{{.Version}}{{end}}{{define "body"}}
{{$n := len .Users}}{{$last := subi $n 1}}{{range $i, $u := .Users}}{{addi $i 1}}/{{$n}} {{if eq $i 0}}*{{else if eq $i $last}}${{else}}-{{end}} {{upper $u.Name}} <{{lower $u.Email}}>{{if $u.Admin}} [ADMIN]{{else if ge $u.Score 60.0}} [PRO]{{else}} [BASIC]{{end}} tags={{join $u.Tags ","}} rev={{join (revstrs $u.Tags) ","}} n={{len $u.Tags}} label={{$u.Label}} disc={{round ($u.Discounted 10.0) 2}} bio={{html $u.Bio}} raw={{$u.Bio}} note={{coalesce $u.Note "none"}}
{{end}}empty=[{{range .Nothing}}x{{else}}vacio{{end}}]
math={{modi (muli (addi (powi 2 5) 4) 3) 100}} abs={{absf -7.5}} concat={{print "go" "-" "veloz"}} logic={{or (and (lt 2 3) (not false)) (gt 1 2)}}
ternary={{if ge 18 18}}adult{{else}}minor{{end}} elvis={{coalesce "" "anon"}} in={{hasint 2 1 2 3}} notin={{not (hasint 9 1 2 3)}}
idx={{atidx 1 10 20 30}} neg={{atidx -1 10 20 30}} city={{index (dict "ciudad" "Barcelona" "pais" "ES") "ciudad"}}
cap={{capitalize "hola mundo"}} title={{titlecase "punto de venta"}} trim=[{{trim "  centrado  "}}] replace={{replace "gato" "g" "p"}} split={{join (split "a;b;c" ";") "+"}}
first={{firstrune "veloz"}} last={{lastrune "veloz"}} nl2br={{nl2br .Salto}}
range={{joinints (seq 1 5) "-"}} max={{maxi 3 8 2}} min={{mini 4 1 7}} len={{runelen "catalunya"}}
keys={{range keys .Meta}}{{.}};{{end}} kv={{range $k, $v := .Meta}}{{$k}}={{$v}};{{end}}
{{$t := 0.0}}{{range .Users}}{{$t = addf $t .Score}}{{end}}total={{$t}}
ws=[{{- if true -}} ok {{- end -}}]
inc=[{{template "footer" .}}]{{end}}REPORT|{{template "body" .}}|END v{{.Version}}`

type goFeatData struct {
	Users   []featUser
	Meta    map[string]string
	Salto   string
	Version string
	Nothing []string
}

const featSalto = "uno\ndos<tres>"

const featVersion = "1.0"

func velozFeatData() map[string]any {
	return map[string]any{
		"users":   featUsers(),
		"meta":    featMeta(),
		"salto":   featSalto,
		"version": featVersion,
		"nothing": []string{},
	}
}

func goFeatDataValue() goFeatData {
	return goFeatData{
		Users:   featUsers(),
		Meta:    featMeta(),
		Salto:   featSalto,
		Version: featVersion,
		Nothing: []string{},
	}
}

func goFeatFuncs() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"join":  strings.Join,
		"trim":  strings.TrimSpace,
		"split": strings.Split,
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"capitalize": func(s string) string {
			if s == "" {
				return s
			}
			r := []rune(s)
			out := make([]rune, len(r))
			out[0] = unicode.ToUpper(r[0])
			for i := 1; i < len(r); i++ {
				out[i] = unicode.ToLower(r[i])
			}
			return string(out)
		},
		"titlecase": func(s string) string {
			var b strings.Builder
			atWordStart := true
			for _, r := range s {
				if unicode.IsLetter(r) || unicode.IsNumber(r) {
					if atWordStart {
						b.WriteRune(unicode.ToUpper(r))
					} else {
						b.WriteRune(unicode.ToLower(r))
					}
					atWordStart = false
				} else {
					b.WriteRune(r)
					atWordStart = true
				}
			}
			return b.String()
		},
		"revstrs": func(xs []string) []string {
			out := make([]string, len(xs))
			for i, x := range xs {
				out[len(xs)-1-i] = x
			}
			return out
		},
		"round": func(f float64, prec int) float64 {
			mult := math.Pow(10, float64(prec))
			return math.Round(f*mult) / mult
		},
		"absf": math.Abs,
		"coalesce": func(s, def string) string {
			if s == "" {
				return def
			}
			return s
		},
		"addi": func(a, b int) int { return a + b },
		"subi": func(a, b int) int { return a - b },
		"muli": func(a, b int) int { return a * b },
		"modi": func(a, b int) int { return a % b },
		"powi": func(a, b int) int {
			out := 1
			for i := 0; i < b; i++ {
				out *= a
			}
			return out
		},
		"addf": func(a, b float64) float64 { return a + b },
		"hasint": func(x int, xs ...int) bool {
			for _, v := range xs {
				if v == x {
					return true
				}
			}
			return false
		},
		"atidx": func(i int, xs ...int) int {
			if i < 0 {
				i += len(xs)
			}
			return xs[i]
		},
		"dict": func(pairs ...string) map[string]string {
			m := make(map[string]string, len(pairs)/2)
			for i := 0; i+1 < len(pairs); i += 2 {
				m[pairs[i]] = pairs[i+1]
			}
			return m
		},
		"firstrune": func(s string) string {
			r := []rune(s)
			if len(r) == 0 {
				return ""
			}
			return string(r[0])
		},
		"lastrune": func(s string) string {
			r := []rune(s)
			if len(r) == 0 {
				return ""
			}
			return string(r[len(r)-1])
		},
		"nl2br": func(s string) string {
			return strings.ReplaceAll(html.EscapeString(s), "\n", "<br />\n")
		},
		"seq": func(start, end int) []int {
			out := make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				out = append(out, i)
			}
			return out
		},
		"joinints": func(xs []int, sep string) string {
			parts := make([]string, len(xs))
			for i, x := range xs {
				parts[i] = strconv.Itoa(x)
			}
			return strings.Join(parts, sep)
		},
		"maxi": func(xs ...int) int {
			best := xs[0]
			for _, x := range xs[1:] {
				if x > best {
					best = x
				}
			}
			return best
		},
		"mini": func(xs ...int) int {
			best := xs[0]
			for _, x := range xs[1:] {
				if x < best {
					best = x
				}
			}
			return best
		},
		"runelen": utf8.RuneCountInString,
		"keys": func(m map[string]string) []string {
			out := make([]string, 0, len(m))
			for k := range m {
				out = append(out, k)
			}
			sort.Strings(out)
			return out
		},
	}
}

func compileVelozFeat(tb testing.TB) *veloz.Template {
	tb.Helper()
	e := veloz.New()
	if _, err := e.Compile("layout", velozLayoutSrc); err != nil {
		tb.Fatal(err)
	}
	if _, err := e.Compile("footer", velozFooterSrc); err != nil {
		tb.Fatal(err)
	}
	tmpl, err := e.Compile("features", velozFeatSrc)
	if err != nil {
		tb.Fatal(err)
	}
	return tmpl
}

func compileGoFeat(tb testing.TB) *texttemplate.Template {
	tb.Helper()
	tmpl, err := texttemplate.New("features").Funcs(goFeatFuncs()).Parse(goFeatSrc)
	if err != nil {
		tb.Fatal(err)
	}
	return tmpl
}

func TestFeatureParity(t *testing.T) {
	vt := compileVelozFeat(t)
	gt := compileGoFeat(t)

	velozOut, err := vt.Render(velozFeatData())
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := gt.Execute(&buf, goFeatDataValue()); err != nil {
		t.Fatal(err)
	}
	goOut := buf.String()

	if velozOut != goOut {
		vLines := strings.Split(velozOut, "\n")
		gLines := strings.Split(goOut, "\n")
		for i := 0; i < len(vLines) || i < len(gLines); i++ {
			var v, g string
			if i < len(vLines) {
				v = vLines[i]
			}
			if i < len(gLines) {
				g = gLines[i]
			}
			if v != g {
				t.Errorf("line %d:\n  veloz: %q\n  go:    %q", i+1, v, g)
			}
		}
	}
}

func BenchmarkFeaturesVeloz(b *testing.B) {
	tmpl := compileVelozFeat(b)
	data := velozFeatData()
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

func BenchmarkFeaturesTextTemplate(b *testing.B) {
	tmpl := compileGoFeat(b)
	data := goFeatDataValue()
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
