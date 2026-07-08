# Roadmap

This is the current direction of the project. It is not a promise: items can move between versions based on feedback and real use. Open an issue if you want to influence the order.

## v0.2

- **Template loaders.** Load templates from a directory or an `fs.FS` (including `embed.FS`), with automatic name resolution for `extends` and `include`. Today every template is compiled by hand with `Compile`.
- **Positions in compile errors.** The lexer and parser already report line numbers; compile errors like "unknown filter" should point to the exact line and column too.
- **Constant folding.** `{{ 2 + 3 * 4 }}` should compile to a single constant instead of five instructions.
- **More built-in filters.** Candidates: `sort`, `map`, `slice`, `batch`, `date`, `number_format`, `trim` variants.
- **Fuzzing in CI, longer runs.** Keep the smoke fuzz on every pull request and add a scheduled job with longer fuzz time.

## v0.3 and later

- **Cached inheritance resolution.** Resolve the extends chain once per template instead of on every render.
- **Specialized iterators.** Iterate native slice types without boxing every element through reflection, removing the remaining per-row allocation.
- **Macros.** Reusable template fragments with parameters, in the spirit of Twig macros.
- **Optional strict mode.** An engine option that turns missing variables and attributes into render errors instead of empty output.

## Explicit non-goals

- **Context-aware autoescaping.** Covering attribute, URL, JS and CSS contexts correctly is a large project of its own; the current model (HTML body escaping, trusted templates) is documented in `SECURITY.md`. Use `html/template` when you need that guarantee.
- **Sandboxed execution of untrusted templates.** Templates are trusted input by design.
