# Benchmarks

All numbers below come from the benchmark suite in this repository. Anyone can reproduce them with one command.

## How to reproduce

```
go test -bench . -benchmem -run '^$'
```

Useful subsets:

```
go test -bench BenchmarkSuite -benchmem -run '^$'      # per-template and scaling comparison
go test -bench BenchmarkScale -benchmem -run '^$'      # dataset scaling 1 to 10k rows
go test -bench BenchmarkFeatures -benchmem -run '^$'   # every engine feature at once
go test -bench BenchmarkCompile -benchmem -run '^$'    # compilation speed
```

## Methodology

- Every comparison renders the same template pair: a Veloz template and its `text/template` equivalent.
- The outputs are byte-identical, enforced by `TestFeatureParity` and `TestGoTemplateSuiteParity`. If the outputs differ, the tests fail, so the benchmarks cannot drift into comparing different work.
- `text/template` gets a `FuncMap` with equivalent helpers where it lacks a built-in (loop metadata, filters). This favors readability, not Veloz: the helpers are simple Go functions.
- Reference machine: Apple M5, Go 1.22, darwin/arm64. Run on your own hardware for your own numbers; ratios are what matter.

## Results

### Per template, fixed data (BenchmarkSuite)

| Template | Veloz | text/template | Ratio |
|---|---|---|---|
| expressions | 1.84 µs · 34 allocs | 8.63 µs · 202 allocs | 4.7x |
| filters | 2.81 µs · 76 allocs | 8.04 µs · 144 allocs | 2.9x |
| functions | 1.46 µs · 39 allocs | 2.94 µs · 66 allocs | 2.0x |
| control | 0.73 µs · 18 allocs | 2.96 µs · 58 allocs | 4.0x |
| loops | 1.53 µs · 30 allocs | 6.38 µs · 116 allocs | 4.2x |
| objects | 1.80 µs · 43 allocs | 2.98 µs · 47 allocs | 1.7x |
| escaping | 0.44 µs · 13 allocs | 1.60 µs · 35 allocs | 3.7x |
| whitespace | 57 ns · 0 allocs | 151 ns · 4 allocs | 2.6x |
| page (inheritance) | 0.99 µs · 13 allocs | 1.40 µs · 18 allocs | 1.4x |

### Scaling, loop template with full loop context

| Rows | Veloz | text/template | Ratio |
|---|---|---|---|
| 1 | 1.16 µs | 3.42 µs | 2.9x |
| 10 | 3.05 µs | 16.8 µs | 5.5x |
| 100 | 21.3 µs | 148 µs | 6.9x |
| 1,000 | 231 µs | 1.56 ms | 6.8x |
| 10,000 | 2.41 ms | 15.1 ms | 6.3x |
| 100,000 | 24.1 ms · 6.4 MB | 149 ms · 81.6 MB | 6.2x, 12.7x less memory |

Part of this gap is expressiveness, not only the VM: `text/template` has no native `loop.revindex` or `loop.last`, so it pays reflection-based `FuncMap` calls on every iteration.

### Scaling, inheritance page (extends + block + include)

| Rows | Veloz | text/template | Ratio |
|---|---|---|---|
| 1 | 0.75 µs | 1.12 µs | 1.5x |
| 100 | 18.0 µs | 25.9 µs | 1.4x |
| 10,000 | 1.71 ms | 2.54 ms | 1.5x |
| 100,000 | 16.8 ms | 25.1 ms | 1.5x |

This is the purest engine-vs-engine comparison: a plain loop plus template composition, no helper functions involved.

### Everything at once (BenchmarkFeatures)

One template that uses filters, inheritance, includes, ternaries, loop context, method calls and autoescape, against a `text/template` equivalent that produces identical output.

| Engine | Time | Allocations |
|---|---|---|
| Veloz | 11.8 µs | 8.4 KB · 189 allocs |
| text/template | 27.4 µs | 12.3 KB · 457 allocs |

### Compilation (BenchmarkCompile)

| Engine | Time |
|---|---|
| Veloz | 3.0 µs |
| text/template | 2.5 µs |

Veloz compiles about 18 percent slower because it does more work up front: local slot allocation, filter binding and bytecode emission. That cost is paid once per template and buys the render speed above.

## Where Veloz loses

Honesty section. With a tiny template and a single data row, both engines finish in under half a microsecond and the difference is noise. `text/template` also compiles slightly faster. If you render templates once and throw them away, Veloz has no advantage.
