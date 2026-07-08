# Roadmap

This is the current direction of the project. It is not a promise: items can move between versions based on feedback and real use. Open an issue if you want to influence the order.

## v0.2 · shipped

- ~~**Template loaders.**~~ `WithFS`, `Load`, lazy cache, `WithReload` and relative paths in `extends` and `include`.
- ~~**Positions in compile errors.**~~ Compile errors report template name, line and column.
- ~~**More built-in filters.**~~ `truncate`, `slice`, `batch`, `sort`, `map`, `date`.

## Next

- **Constant folding.** `{{ 2 + 3 * 4 }}` should compile to a single constant instead of five instructions.
- **Fuzzing in CI, longer runs.** Keep the smoke fuzz on every pull request and add a scheduled job with longer fuzz time.

## v0.3 and later

- **Positions in render errors.** Map bytecode instructions back to template positions so runtime failures point at the template line too.

- **Cached inheritance resolution.** Resolve the extends chain once per template instead of on every render.
- **Specialized iterators.** Iterate native slice types without boxing every element through reflection, removing the remaining per-row allocation.
- **Macros.** Reusable template fragments with parameters, in the spirit of Twig macros.
- **Optional strict mode.** An engine option that turns missing variables and attributes into render errors instead of empty output.

## Explicit non-goals

- **Context-aware autoescaping.** Covering attribute, URL, JS and CSS contexts correctly is a large project of its own; the current model (HTML body escaping, trusted templates) is documented in `SECURITY.md`. Use `html/template` when you need that guarantee.
- **Sandboxed execution of untrusted templates.** Templates are trusted input by design.
