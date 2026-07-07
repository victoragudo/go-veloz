# Veloz

[![CI](https://github.com/victoragudo/go-veloz/actions/workflows/ci.yml/badge.svg)](https://github.com/victoragudo/go-veloz/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/victoragudo/go-veloz.svg)](https://pkg.go.dev/github.com/victoragudo/go-veloz)
[![License: MIT](https://img.shields.io/badge/License-MIT-red.svg)](LICENSE)

A fast template engine for Go with a Twig/Blade style syntax. Templates compile once to compact bytecode and run on a stack virtual machine, up to 7x faster than `text/template` with 12x fewer allocations.

**[Website](https://victoragudo.github.io/go-veloz/) · [Documentation](https://victoragudo.github.io/go-veloz/docs.html)**

## Install

```
go get github.com/victoragudo/go-veloz
```

## Quick start

```go
package main

import (
    "fmt"

    veloz "github.com/victoragudo/go-veloz"
)

func main() {
    engine := veloz.New()
    tmpl, err := engine.Compile("hello", "Hola {{ name | capitalize }}!")
    if err != nil {
        panic(err)
    }
    out, err := tmpl.Render(map[string]any{"name": "mireia"})
    if err != nil {
        panic(err)
    }
    fmt.Println(out)
}
```

## A taste of the syntax

```twig
{% extends "layout" %}
{% block body %}
{% for line in invoice.lines %}
  {{ loop.index }}. {{ line.concept | capitalize }} = {{ line.total() | money }}
{% endfor %}
Status: {{ invoice.paid ? "PAID" : "PENDING" }}
Client: {{ client ?: "guest" }}
{% endblock %}
```

## Features

- Twig style expressions: ternary, elvis `?:`, `in`, string concat `~`, power `**`, negative indexing, array and map literals
- Template inheritance with `extends` and `block`, plus `include` for partials
- Real loop context: `loop.index`, `first`, `last`, `revindex`, `length`, key-value iteration and `for/else`
- 20 built-in filters, custom filters and functions in one call
- HTML autoescape by default, with `raw` and `SafeString` escape hatches
- Compile time checks: unknown filters and broken tags fail at `Compile`, not in production
- Zero dependencies, safe for concurrent use, pooled interpreters with no fixed allocations per render

## Performance

Rendering the same templates with byte-identical output, Apple M5, Go 1.22:

| Scenario | Veloz | text/template | Speedup |
|---|---|---|---|
| Loop with loop context, 100 rows | 21.3 µs | 147.7 µs | 6.9x |
| Escaped HTML output | 0.44 µs | 1.60 µs | 3.7x |
| Full feature template | 11.8 µs | 27.4 µs | 2.3x |
| Inheritance page, 100k rows | 16.8 ms | 25.1 ms | 1.5x |

Run them yourself with `go test -bench . -benchmem`.

## Documentation

The full reference lives at [victoragudo.github.io/go-veloz/docs.html](https://victoragudo.github.io/go-veloz/docs.html): expressions, filters, functions, loops, inheritance, autoescape and the Go API, each with examples.

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

## License

[MIT](LICENSE)
