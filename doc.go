// Package veloz is a fast, flexible template engine for Go with a Twig/Blade-style
// syntax.
//
// Templates are compiled to a compact bytecode and executed on a stack virtual
// machine. Path resolution, filter binding, and control-flow structure are resolved
// once at compile time instead of being re-walked on every render, so templates load
// at runtime while rendering faster than the standard library's text/template.
//
// Basic usage:
//
//	engine := veloz.New()
//	tmpl, err := engine.Compile("hello", "Hola {{ name | capitalize }}!")
//	if err != nil {
//		// handle error
//	}
//	out, err := tmpl.Render(map[string]any{"name": "mireia"})
//
// The engine supports expressions, filters and functions (sharing one namespace),
// if/elseif/else, for loops with a loop context, set, automatic HTML escaping,
// template inheritance via extends/block, and include. Register custom behaviour with
// RegisterFilter and RegisterFunction; expose custom attributes on your own types by
// implementing Attributer.
package veloz
