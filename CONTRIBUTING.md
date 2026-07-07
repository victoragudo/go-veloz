# Contributing to Veloz

Thank you for helping to improve Veloz. This guide keeps contributions simple and fast to review.

## How to contribute

1. Fork the repository and create a branch from `main`.
2. Make your change. Keep it focused: one topic per pull request.
3. Add or update tests. Every feature and bug fix needs test coverage.
4. Make sure everything passes locally:

```
go build ./...
go vet ./...
go test ./...
gofmt -l .
```

5. Open a pull request against `main` with a clear description of the problem and the solution.

`main` is a protected branch: changes only land through pull requests, CI must be green and a maintainer review is required.

## Style

- Run `gofmt` before committing.
- Follow the existing code style: no code comments, clear names, small functions.
- Template behavior changes need a test in `veloz_test.go` or a `.tpl` case in `testdata/` with its golden file.
- Performance sensitive changes should include benchmark numbers before and after.

## Reporting bugs

Open an issue with a minimal template that reproduces the problem, the data you passed, the output you got and the output you expected.

## Commit messages

Write short, clear messages in English that explain the intent of the change.
