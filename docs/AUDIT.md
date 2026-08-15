# Project audit and improvement plan

## Current state

- The application is a Go/Fyne desktop GUI for RAM and CPU stress testing.
- The previous implementation mixed UI, localization, orchestration, and stress-test logic in root-level files, which made future changes harder to review and test.
- Build and verification commands were documented in the README but were not available as repeatable Make targets.
- Dependency updates require network access to the Go module proxy or upstream Git remotes.

## Improvement plan

1. **Architecture and developer ergonomics** — separate UI, localization, and stress-test engine code into focused packages; keep `main.go` as a thin entry point; add a `Makefile` for common workflows.
2. **Correctness and safety** — add focused unit tests for non-GUI logic, make event delivery consistently non-blocking, and keep cancellation paths explicit.
3. **Dependency maintenance** — periodically run `make update` in an environment with access to `proxy.golang.org` or upstream GitHub remotes, then commit `go.mod` and `go.sum` changes.
4. **CI validation** — ensure Linux runners install Fyne native dependencies (`OpenGL`, `X11`, `Xcursor`, etc.) before `go test`/`go build`.
5. **Documentation** — document required native dependencies and recommended safe stress-test limits for end users.

## Audit notes from this pass

- The first plan item has been implemented: the GUI now lives in `internal/ui`, localization in `internal/i18n`, and stress-test execution in `internal/stress`.
- A `Makefile` now provides `fmt`, `test`, `build`, `run`, `tidy`, `update`, `clean`, and `audit` targets.
- Dependency update was attempted, but this container cannot reach the Go proxy or GitHub remotes (`403` responses). No dependency versions were changed.
- Local correctness verification is partially blocked by missing native Fyne build dependencies in the container (`gl.pc` and `X11/Xcursor/Xcursor.h`).
