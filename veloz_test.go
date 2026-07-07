package veloz_test

import (
	"strconv"
	"testing"

	"github.com/victoragudo/go-veloz"
)

type product struct {
	Name  string
	Price float64
	Tags  []string
}

func (p product) Label() string {
	return p.Name + " (" + strconv.FormatFloat(p.Price, 'f', 2, 64) + ")"
}

func (p product) Discounted(pct float64) float64 {
	return p.Price * (1 - pct/100)
}

func render(t *testing.T, src string, data any) string {
	t.Helper()
	e := veloz.New()
	tmpl, err := e.Compile("test", src)
	if err != nil {
		t.Fatalf("compile error for %q: %v", src, err)
	}
	out, err := tmpl.Render(data)
	if err != nil {
		t.Fatalf("render error for %q: %v", src, err)
	}
	return out
}

func TestExpressions(t *testing.T) {
	cases := []struct {
		name string
		src  string
		data any
		want string
	}{
		{"text", "hola mundo", nil, "hola mundo"},
		{"int", "{{ 42 }}", nil, "42"},
		{"float", "{{ 3.5 }}", nil, "3.5"},
		{"string", `{{ "hola" }}`, nil, "hola"},
		{"bool_true", "{{ true }}", nil, "true"},
		{"nil", "{{ null }}", nil, ""},
		{"add", "{{ 2 + 3 }}", nil, "5"},
		{"precedence", "{{ 2 + 3 * 4 }}", nil, "14"},
		{"parens", "{{ (2 + 3) * 4 }}", nil, "20"},
		{"div_float", "{{ 10 / 4 }}", nil, "2.5"},
		{"mod", "{{ 10 % 3 }}", nil, "1"},
		{"pow", "{{ 2 ** 10 }}", nil, "1024"},
		{"neg", "{{ -5 + 2 }}", nil, "-3"},
		{"concat", `{{ "Hola " ~ "Mireia" }}`, nil, "Hola Mireia"},
		{"eq_true", "{{ 3 == 3 }}", nil, "true"},
		{"lt", "{{ 2 < 5 }}", nil, "true"},
		{"gte", "{{ 5 >= 5 }}", nil, "true"},
		{"string_cmp", `{{ "abc" < "abd" }}`, nil, "true"},
		{"and_operand", `{{ true and "yes" }}`, nil, "yes"},
		{"or_fallback", `{{ false or "fallback" }}`, nil, "fallback"},
		{"zero_or", `{{ 0 or "x" }}`, nil, "x"},
		{"not", "{{ not false }}", nil, "true"},
		{"ternary", `{{ 20 >= 18 ? "adult" : "minor" }}`, nil, "adult"},
		{"elvis", `{{ "" ?: "anon" }}`, nil, "anon"},
		{"elvis_keep", `{{ "Oriol" ?: "anon" }}`, nil, "Oriol"},
		{"in_string", `{{ "a" in "cat" }}`, nil, "true"},
		{"in_list", "{{ 2 in [1, 2, 3] }}", nil, "true"},
		{"not_in", "{{ 9 not in [1, 2, 3] }}", nil, "true"},
		{"array_index", "{{ [10, 20, 30][1] }}", nil, "20"},
		{"array_neg_index", "{{ [10, 20, 30][-1] }}", nil, "30"},
		{"map_literal", `{{ {name: "Aleix", age: 30}.name }}`, nil, "Aleix"},
		{"map_index", `{{ {"clave": "valor"}["clave"] }}`, nil, "valor"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := render(t, tc.src, tc.data); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVariablesAndAccess(t *testing.T) {
	data := map[string]any{
		"name": "Mireia",
		"user": map[string]any{
			"email": "mireia@example.com",
			"roles": []string{"admin", "editor"},
		},
		"prod": product{Name: "Teclado", Price: 49.9, Tags: []string{"perifericos", "oferta"}},
	}
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"var", "{{ name }}", "Mireia"},
		{"nested_attr", "{{ user.email }}", "mireia@example.com"},
		{"nested_index", "{{ user.roles[0] }}", "admin"},
		{"struct_field", "{{ prod.Name }}", "Teclado"},
		{"struct_float", "{{ prod.Price }}", "49.9"},
		{"struct_slice", "{{ prod.Tags[1] }}", "oferta"},
		{"method_call", "{{ prod.Label() }}", "Teclado (49.90)"},
		{"method_args", "{{ prod.Discounted(10) }}", "44.91"},
		{"missing_lenient", "{{ user.missing }}", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := render(t, tc.src, data); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFilters(t *testing.T) {
	data := map[string]any{
		"names": []string{"Mireia", "Aleix", "Nuria"},
		"price": 3.14159,
		"empty": "",
	}
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"upper", `{{ "hola" | upper }}`, "HOLA"},
		{"lower", `{{ "HOLA" | lower }}`, "hola"},
		{"capitalize", `{{ "hola mundo" | capitalize }}`, "Hola mundo"},
		{"title", `{{ "hola mundo cruel" | title }}`, "Hola Mundo Cruel"},
		{"trim", `{{ "  hola  " | trim }}`, "hola"},
		{"length", "{{ names | length }}", "3"},
		{"join", `{{ names | join(", ") }}`, "Mireia, Aleix, Nuria"},
		{"join_chain", `{{ names | reverse | join("-") }}`, "Nuria-Aleix-Mireia"},
		{"first", "{{ names | first }}", "Mireia"},
		{"last", "{{ names | last }}", "Nuria"},
		{"round", "{{ price | round(2) }}", "3.14"},
		{"default", "{{ empty | default(\"sin nombre\") }}", "sin nombre"},
		{"default_keep", `{{ "Oriol" | default("x") }}`, "Oriol"},
		{"replace", `{{ "gato" | replace("g", "p") }}`, "pato"},
		{"function_form", `{{ upper("hola") }}`, "HOLA"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := render(t, tc.src, data); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFunctions(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"{{ range(1, 5) | join(\",\") }}", "1,2,3,4,5"},
		{"{{ max(3, 8, 2) }}", "8"},
		{"{{ min([4, 1, 7]) }}", "1"},
		{"{{ length([1, 2, 3, 4]) }}", "4"},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			if got := render(t, tc.src, nil); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestControlFlow(t *testing.T) {
	cases := []struct {
		name string
		src  string
		data any
		want string
	}{
		{
			"if_true",
			`{% if age >= 18 %}adult{% endif %}`,
			map[string]any{"age": 25},
			"adult",
		},
		{
			"if_else",
			`{% if age >= 18 %}adult{% else %}minor{% endif %}`,
			map[string]any{"age": 12},
			"minor",
		},
		{
			"elseif",
			`{% if n < 0 %}neg{% elseif n == 0 %}zero{% else %}pos{% endif %}`,
			map[string]any{"n": 0},
			"zero",
		},
		{
			"for",
			`{% for name in names %}{{ name }};{% endfor %}`,
			map[string]any{"names": []string{"Mireia", "Aleix"}},
			"Mireia;Aleix;",
		},
		{
			"for_loop_index",
			`{% for name in names %}{{ loop.index }}:{{ name }}{% if not loop.last %}, {% endif %}{% endfor %}`,
			map[string]any{"names": []string{"Mireia", "Aleix", "Nuria"}},
			"1:Mireia, 2:Aleix, 3:Nuria",
		},
		{
			"for_else_empty",
			`{% for x in items %}{{ x }}{% else %}vacio{% endfor %}`,
			map[string]any{"items": []int{}},
			"vacio",
		},
		{
			"for_key_value",
			`{% for k, v in scores %}{{ k }}={{ v }} {% endfor %}`,
			map[string]any{"scores": map[string]int{"aleix": 90, "mireia": 95}},
			"aleix=90 mireia=95 ",
		},
		{
			"nested_for",
			`{% for row in grid %}{% for cell in row %}{{ cell }}{% endfor %}|{% endfor %}`,
			map[string]any{"grid": [][]int{{1, 2}, {3, 4}}},
			"12|34|",
		},
		{
			"set_accumulator",
			`{% set total = 0 %}{% for n in nums %}{% set total = total + n %}{% endfor %}{{ total }}`,
			map[string]any{"nums": []int{1, 2, 3, 4}},
			"10",
		},
		{
			"set_simple",
			`{% set greeting = "Hola " ~ name %}{{ greeting }}`,
			map[string]any{"name": "Nuria"},
			"Hola Nuria",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := render(t, tc.src, tc.data); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoopStateCapture(t *testing.T) {
	src := `{% set snap = null %}{% for x in items %}{% if loop.first %}{% set snap = loop %}{% endif %}{% endfor %}{{ snap.index }}/{{ snap.length }}`
	data := map[string]any{"items": []string{"norte", "sur", "este"}}
	if got := render(t, src, data); got != "1/3" {
		t.Errorf("captured loop state mutated: got %q, want %q", got, "1/3")
	}
}

func TestModuloFloat(t *testing.T) {
	if got := render(t, "{{ 10.5 % 3 }}", nil); got != "1.5" {
		t.Errorf("got %q, want %q", got, "1.5")
	}
}

func TestBoolNumericEquality(t *testing.T) {
	if got := render(t, "{{ true == 1 }}", nil); got != "true" {
		t.Errorf("got %q, want %q", got, "true")
	}
	if got := render(t, "{{ false == 0 }}", nil); got != "true" {
		t.Errorf("got %q, want %q", got, "true")
	}
}

func TestAutoescape(t *testing.T) {
	data := map[string]any{"html": `<b>peligro & "comillas"</b>`}
	if got := render(t, "{{ html }}", data); got != "&lt;b&gt;peligro &amp; &#34;comillas&#34;&lt;/b&gt;" {
		t.Errorf("autoescape failed: %q", got)
	}
	if got := render(t, "{{ html | raw }}", data); got != `<b>peligro & "comillas"</b>` {
		t.Errorf("raw failed: %q", got)
	}
}

func TestAutoescapeDisabled(t *testing.T) {
	e := veloz.New(veloz.WithAutoescape(false))
	tmpl, err := e.Compile("t", "{{ html }}")
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(map[string]any{"html": "<b>x</b>"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "<b>x</b>" {
		t.Errorf("got %q", out)
	}
}

func TestInclude(t *testing.T) {
	e := veloz.New()
	if _, err := e.Compile("partial", `Hola {{ name }}`); err != nil {
		t.Fatal(err)
	}
	tmpl, err := e.Compile("main", `[{% include "partial" %}]`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(map[string]any{"name": "Mireia"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "[Hola Mireia]" {
		t.Errorf("got %q", out)
	}
}

func TestInheritance(t *testing.T) {
	e := veloz.New()
	if _, err := e.Compile("base", `<page><h1>{% block title %}Base{% endblock %}</h1><div>{% block body %}empty{% endblock %}</div></page>`); err != nil {
		t.Fatal(err)
	}
	tmpl, err := e.Compile("child", `{% extends "base" %}{% block title %}Perfil de Aleix{% endblock %}{% block body %}Contenido{% endblock %}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(nil)
	if err != nil {
		t.Fatal(err)
	}
	want := `<page><h1>Perfil de Aleix</h1><div>Contenido</div></page>`
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestInheritancePartialOverride(t *testing.T) {
	e := veloz.New()
	if _, err := e.Compile("layout", `[{% block header %}default-header{% endblock %}][{% block footer %}default-footer{% endblock %}]`); err != nil {
		t.Fatal(err)
	}
	tmpl, err := e.Compile("view", `{% extends "layout" %}{% block header %}custom{% endblock %}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tmpl.Render(nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "[custom][default-footer]" {
		t.Errorf("got %q", out)
	}
}

func TestCompileErrors(t *testing.T) {
	e := veloz.New()
	cases := []string{
		`{{ unknownFilterHere("x") | nonexistentfilter }}`,
		`{% if x %}no end`,
		`{{ 1 + }}`,
		`{% for x %}{% endfor %}`,
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			if _, err := e.Compile("bad", src); err == nil {
				t.Errorf("expected compile error for %q", src)
			}
		})
	}
}

func TestWhitespaceControl(t *testing.T) {
	src := "a\n{%- if true -%}\n  b\n{%- endif -%}\nc"
	got := render(t, src, nil)
	if got != "abc" {
		t.Errorf("got %q, want %q", got, "abc")
	}
}
