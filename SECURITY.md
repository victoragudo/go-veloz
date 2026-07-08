# Security model

## Trust boundaries

Veloz follows the same trust model as most template engines outside the Go standard library (Twig, Jinja, Blade):

- **Templates are trusted input.** They are code written by developers, like your Go source. Do not compile templates that come from end users. A hostile template author can produce large outputs or slow renders on purpose.
- **Data is untrusted input.** Values passed to `Render` are escaped by default and can come from users, databases or external APIs.

## What autoescape covers

With autoescape on (the default), every value printed with `{{ }}` is HTML-escaped: `& < > " '` become entities. This protects the common case of interpolating untrusted data into HTML element content.

```twig
<p>{{ user_comment }}</p>   safe
```

## What autoescape does NOT cover

Escaping is for HTML body context only. It is **not context-aware** like Go's `html/template`. The engine does not know when a value lands inside an attribute, a URL, inline JavaScript or CSS. These are unsafe with untrusted data:

```twig
<a href="{{ user_url }}">          unsafe: javascript: URLs pass through
<div onclick="go('{{ name }}')">   unsafe: JS string context
<script>var x = {{ user_json }};</script>   unsafe
```

For those cases, validate or encode the value in Go before passing it to the template. If you need strict context-aware escaping of untrusted data in every position, use `html/template`.

## Escape hatches

Three ways to bypass escaping, all explicit:

- the `raw` filter: `{{ trusted_html | raw }}`
- returning a `veloz.SafeString` from a custom filter or function
- `veloz.New(veloz.WithAutoescape(false))` for non-HTML output such as config files or code generation

Anything marked safe is written verbatim. Only mark values you control.

## Robustness

- Unknown filters, unknown functions and malformed tags fail at `Compile` time.
- The parser limits nesting depth, and `range()` limits its result size, so a malformed or hostile template fails with an error instead of exhausting the stack or memory.
- The compiler and runtime are continuously fuzzed (`go test -fuzz FuzzCompile`, `go test -fuzz FuzzCompileAndRender`).
- Missing variables render as empty strings and never panic.

## Reporting a vulnerability

Open a private security advisory on GitHub (Security tab, "Report a vulnerability") or open an issue if the report is not sensitive. Please include a minimal template and data that reproduce the problem.
