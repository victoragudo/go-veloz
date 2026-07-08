package main

import (
	"os"

	veloz "github.com/victoragudo/go-veloz"
)

const emailSrc = `Subject: Your order {{ order.id }} has shipped

Hola {{ customer.name | capitalize }},

Good news: your order is on the way.

{% for item in order.items -%}
  - {{ item.qty }}x {{ item.name }}
{% endfor -%}
Carrier: {{ order.carrier | default("standard shipping") }}
Tracking: {{ order.tracking ?: "available soon" }}

Thanks for shopping with us,
{{ shop }}
`

func main() {
	engine := veloz.New(veloz.WithAutoescape(false))
	tmpl := engine.MustCompile("shipped", emailSrc)

	data := map[string]any{
		"shop":     "Tienda Central",
		"customer": map[string]any{"name": "mireia"},
		"order": map[string]any{
			"id":       "2026-0142",
			"carrier":  "",
			"tracking": "ES7712834455",
			"items": []map[string]any{
				{"qty": 1, "name": "Mechanical keyboard"},
				{"qty": 2, "name": "USB-C cable"},
			},
		},
	}

	if err := tmpl.RenderTo(os.Stdout, data); err != nil {
		panic(err)
	}
}
