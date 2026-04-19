# Repository Guidelines

## Project Structure & Module Organization
This repository is currently a small single-module Go application. `main.go` is the CLI entrypoint and contains the runnable program. `go.mod` is the source of truth for the module name and Go version (`go 1.26`). The `.idea/` directory contains local IDE metadata and should not drive project structure decisions. As the codebase grows, move reusable logic out of `main.go` into focused `internal/...` packages and keep related tests beside the package they cover.

## Build, Test, and Development Commands
Run `go run .` to execute the app locally. Run `go build ./...` to compile all packages and catch build issues before review. Run `go test ./...` to execute the test suite; at the moment this mostly acts as a compile and vet check because no `_test.go` files exist yet. Run `go fmt ./...` before committing to apply standard Go formatting.

## Coding Style & Naming Conventions
Follow idiomatic Go. Let `gofmt` control whitespace and indentation; do not hand-align code. Use `CamelCase` for exported names, `camelCase` for unexported helpers, and short package names that match their directory. Keep `main()` thin: parse inputs, call package-level logic, and print results. Prefer descriptive file names such as `album_store.go` or `album_store_test.go`.

## Testing Guidelines
Use Go's built-in `testing` package. Add table-driven tests for business logic and name test files `*_test.go` with functions like `TestLoadAlbum` or `TestRenderPhotoList`. New behavior should include tests even though no formal coverage threshold exists yet.

## Commit & Pull Request Guidelines
No Git history is available in this workspace, so use these as recommended defaults until the repository establishes its own convention. Write commit messages in Conventional Commit style, for example `feat: add album loader` or `fix: correct CLI greeting`. Pull requests should include a short summary, linked issue when applicable, and the verification commands you ran, such as `go fmt ./...` and `go test ./...`.
