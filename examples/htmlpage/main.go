package main

import (
	"os"

	veloz "github.com/victoragudo/go-veloz"
)

const layoutSrc = `<!doctype html>
<html>
<head><title>{% block title %}Shop{% endblock %}</title></head>
<body>
<h1>{% block title %}Shop{% endblock %}</h1>
{% block body %}{% endblock %}
<footer>{{ shop }} · {{ year }}</footer>
</body>
</html>
`

const pageSrc = `{% extends "layout" %}
{% block title %}Catalog{% endblock %}
{% block body %}
<ul>
{%- for p in products %}
  <li>{{ p.name }} - {{ p.price }} EUR{% if p.featured %} <strong>featured</strong>{% endif %}</li>
{%- endfor %}
</ul>
<p>Search note: {{ user_note }}</p>
{% endblock %}`

func main() {
	engine := veloz.New()
	engine.MustCompile("layout", layoutSrc)
	page := engine.MustCompile("catalog", pageSrc)

	data := map[string]any{
		"shop": "Tienda Central",
		"year": 2026,
		"products": []map[string]any{
			{"name": "Mechanical keyboard", "price": 89.9, "featured": true},
			{"name": "Vertical mouse", "price": 45.5, "featured": false},
		},
		"user_note": `<script>alert("this gets escaped")</script>`,
	}

	if err := page.RenderTo(os.Stdout, data); err != nil {
		panic(err)
	}
}
