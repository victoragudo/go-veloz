package veloz_test

import (
	"testing"

	veloz "github.com/victoragudo/go-veloz"
)

func fuzzSeeds() []string {
	return []string{
		"hola mundo",
		"{{ name }}",
		"{{ 2 + 3 * 4 }}",
		"{{ user.email | upper | default(\"x\") }}",
		"{% if age >= 18 %}adult{% elseif age > 0 %}minor{% else %}?{% endif %}",
		"{% for k, v in stock %}{{ loop.index }}:{{ k }}={{ v }}{% else %}empty{% endfor %}",
		"{% set total = 0 %}{% for p in prices %}{% set total = total + p %}{% endfor %}{{ total }}",
		"{% extends \"layout\" %}{% block body %}{{ name }}{% endblock %}",
		"{% include \"partial\" %}",
		"{{ [10, 20, 30][-1] }} {{ {city: \"Madrid\"}.city }}",
		"{{ 1 ? \"a\" : \"b\" }} {{ \"\" ?: \"anon\" }} {{ 2 in [1, 2] }}",
		"{#- comment -#}{{- \"x\" -}}",
		"{{ range(1, 5) | join(\",\") }} {{ max(3, 8) }}",
		"{{ prod.Label() }} {{ prod.Discounted(10) | round(2) }}",
		"{{ \"a\" ~ \"b\" ~ 3 ** 2 % 5 }}",
		"{{ ((((1)))) }}",
		"{% if x %}{% if y %}{% if z %}deep{% endif %}{% endif %}{% endif %}",
		"{{ html | raw }} {{ html | escape }} {{ notes | nl2br }}",
		"{{",
		"{%",
		"{% for %}",
		"{{ 1 + }}",
		"{% endif %}",
		"{{ \"unterminated }}",
	}
}

func fuzzData() map[string]any {
	return map[string]any{
		"name":   "Mireia",
		"age":    30,
		"x":      true,
		"y":      1,
		"z":      "yes",
		"user":   map[string]any{"email": "mireia@example.com"},
		"stock":  map[string]int{"keyboard": 12, "mouse": 30},
		"prices": []float64{19.9, 45, 5.5},
		"html":   `<b>bold & "quoted"</b>`,
		"notes":  "one\ntwo",
		"prod":   product{Name: "Teclado", Price: 49.9},
	}
}

func FuzzCompile(f *testing.F) {
	for _, seed := range fuzzSeeds() {
		f.Add(seed)
	}
	for _, name := range append(append([]string{}, sharedTemplates...), suiteTemplates...) {
		f.Add(readTemplateFile(f, name))
	}
	f.Fuzz(func(t *testing.T, src string) {
		e := veloz.New()
		if _, err := e.Compile("fuzz", src); err != nil {
			return
		}
	})
}

func FuzzCompileAndRender(f *testing.F) {
	for _, seed := range fuzzSeeds() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		if len(src) > 4096 {
			return
		}
		e := veloz.New()
		if _, err := e.Compile("layout", "L{% block body %}B{% endblock %}"); err != nil {
			return
		}
		if _, err := e.Compile("partial", "P{{ name }}"); err != nil {
			return
		}
		tmpl, err := e.Compile("fuzz", src)
		if err != nil {
			return
		}
		_, _ = tmpl.Render(fuzzData())
	})
}
