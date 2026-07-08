# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/) and the project uses [Semantic Versioning](https://semver.org/).

## v0.1.2 · 2026-07-08

### Added

- Fuzz tests for the compiler and the runtime (`FuzzCompile`, `FuzzCompileAndRender`), also running as a smoke check in CI.
- Runnable examples in `examples/`: HTML page, transactional email, invoice, nginx config generation.
- `BENCHMARKS.md` with the full benchmark tables in text and instructions to reproduce them.
- `SECURITY.md` describing the trust model and the exact scope of autoescape.

### Fixed

- Calling a built-in filter with no arguments, like `{{ round() }}`, returned an error instead of panicking. Found by fuzzing.
- The parser now limits nesting depth, so deeply nested input fails with a compile error instead of exhausting the stack.
- `range()` now limits its result to one million elements instead of allocating without bounds.

## v0.1.1 · 2026-07-07

Same code as v0.1.0, republished with corrected commit author metadata. Use this version instead of v0.1.0.

## v0.1.0 · 2026-07-07

First public release.

- Twig style template language: expressions, ternary and elvis operators, `in`, string concat `~`, array and map literals, negative indexing.
- Control flow: `if` / `elseif` / `else`, `set`, `for` with full loop context and `for/else`.
- Template inheritance with `extends` and `block`, includes, whitespace control and comments.
- 20 built-in filters, 4 functions, custom filters and functions through one registration call.
- HTML autoescape by default with `raw`, `escape` and `SafeString`.
- Bytecode compiler and stack VM with pooled, reusable execution frames.
