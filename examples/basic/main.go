package main

import (
	"fmt"
	"os"
	"time"

	"veloz"
)

type invoice struct {
	Number string
	Lines  []line
	Paid   bool
}

type line struct {
	Concept  string
	Quantity int
	Price    float64
}

func (l line) Total() float64 {
	return float64(l.Quantity) * l.Price
}

func main() {
	engine := veloz.New()

	engine.RegisterFilter("money", func(args []veloz.Value) (veloz.Value, error) {
		f, _ := args[0].Interface().(float64)
		if n, ok := args[0].Interface().(int64); ok {
			f = float64(n)
		}
		return veloz.Str(fmt.Sprintf("%.2f€", f)), nil
	})

	engine.RegisterFunction("today", func(args []veloz.Value) (veloz.Value, error) {
		return veloz.Str(time.Now().Format("2006-01-02")), nil
	})

	if _, err := engine.Compile("layout", layoutSrc); err != nil {
		panic(err)
	}
	tmpl, err := engine.Compile("invoice", invoiceSrc)
	if err != nil {
		panic(err)
	}

	data := map[string]any{
		"invoice": invoice{
			Number: "2026-014",
			Paid:   false,
			Lines: []line{
				{Concept: "Diseño", Quantity: 3, Price: 120.0},
				{Concept: "Desarrollo", Quantity: 10, Price: 85.5},
				{Concept: "Soporte", Quantity: 1, Price: 200.0},
			},
		},
		"client": "Estudio Marés",
	}

	if err := tmpl.RenderTo(os.Stdout, data); err != nil {
		panic(err)
	}
}

const layoutSrc = `================ FACTURA ================
{% block body %}{% endblock %}
Generado el {{ today() }}
=========================================
`

const invoiceSrc = `{% extends "layout" %}
{% block body %}
Factura Nº {{ invoice.number }} — Cliente: {{ client }}
Estado: {% if invoice.paid %}PAGADA{% else %}PENDIENTE{% endif %}

{% for l in invoice.lines %}  {{ loop.index }}. {{ l.concept }} x{{ l.quantity }} @ {{ l.price | money }} = {{ l.total() | money }}
{% endfor %}
{% set total = 0 %}{% for l in invoice.lines %}{% set total = total + l.total() %}{% endfor %}TOTAL: {{ total | money }}
{% endblock %}`
