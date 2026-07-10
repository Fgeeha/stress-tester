# Stress Tester

This project provides a cross-platform memory and CPU stress testing tool written in Go and built with the [Fyne](https://fyne.io/) GUI framework. Use it to verify hardware stability under heavy load.

## Building

You can build the application using the Go toolchain or the provided Makefile. The project is tested with Go 1.21.

```bash
# build for your current platform
make build

# or use Go directly
go build -o build/stress-tester
```

To run the application directly without building a binary:

```bash
make run
# or
go run .
```

Useful maintenance commands:

```bash
make fmt     # format Go code
make test    # run Go tests / compile packages
make update  # update Go modules and tidy go.mod/go.sum
make audit   # format and run tests
```

## Usage

Launch the application and select the amount of memory and number of threads to stress. Press **Start** to begin the test and **Stop** to end it. Logs are written to `stress_test.log` in the project directory.

### Memory Safety

The tester monitors RAM and swap usage while running. If both become critically low, the stress test stops automatically to prevent crashes. Memory allocation failures are logged and no longer hang the application.

## Architecture

The entry point is intentionally small and delegates to focused internal packages:

- `internal/ui` contains Fyne window construction, event handling, and test orchestration.
- `internal/stress` contains CPU and memory stress-test execution.
- `internal/i18n` contains UI translations.
- `logger` contains log-file and UI event helpers.

See `docs/AUDIT.md` for the current audit notes and improvement plan.

## GitHub Actions

A workflow file in `.github/workflows/go-build.yml` builds the project automatically on push and pull request for Ubuntu and Windows runners.

## Disclaimer

This tool is intended for legitimate equipment testing only. Do not use it for malicious purposes.
