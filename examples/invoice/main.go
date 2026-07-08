package main

import (
	"fmt"
	"os"
	"time"

	veloz "github.com/victoragudo/go-veloz"
)

type line struct {
	Concept  string
	Quantity int
	Price    float64
}

func (l line) Total() float64 {
	return float64(l.Quantity) * l.Price
}

const invoiceSrc = `================ INVOICE {{ number }} ================
Client: {{ client }}
Date:   {{ today() }}

{% for l in lines %}  {{ loop.index }}. {{ l.Concept }} x{{ l.Quantity }} @ {{ l.Price | money }} = {{ l.Total() | money }}
{% endfor %}
{%- set total = 0 %}{% for l in lines %}{% set total = total + l.Total() %}{% endfor %}
TOTAL: {{ total | money }}
Status: {{ paid ? "PAID" : "PENDING" }}
`

func main() {
	engine := veloz.New(veloz.WithAutoescape(false))

	engine.RegisterFilter("money", func(args []veloz.Value) (veloz.Value, error) {
		f, ok := args[0].Interface().(float64)
		if !ok {
			if n, isInt := args[0].Interface().(int64); isInt {
				f = float64(n)
			}
		}
		return veloz.Str(fmt.Sprintf("%.2f EUR", f)), nil
	})

	engine.RegisterFunction("today", func(args []veloz.Value) (veloz.Value, error) {
		return veloz.Str(time.Now().Format("2006-01-02")), nil
	})

	tmpl := engine.MustCompile("invoice", invoiceSrc)

	data := map[string]any{
		"number": "2026-014",
		"client": "Estudio Mares",
		"paid":   false,
		"lines": []line{
			{Concept: "Design", Quantity: 3, Price: 120},
			{Concept: "Development", Quantity: 10, Price: 85.5},
			{Concept: "Support", Quantity: 1, Price: 200},
		},
	}

	if err := tmpl.RenderTo(os.Stdout, data); err != nil {
		panic(err)
	}
}
