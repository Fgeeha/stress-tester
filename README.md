# Stress Tester

This project provides a cross-platform memory and CPU stress testing tool written in Go and built with the [Fyne](https://fyne.io/) GUI framework. Use it to verify hardware stability under heavy load.

## Building

You can build the application using the Go toolchain. The project is tested with Go 1.21.

```bash
# build for your current platform
go build -o stress-tester
```

To run the application directly without building a binary:

```bash
go run .
```

## Usage

Launch the application and select the amount of memory and number of threads to stress. Press **Start** to begin the test and **Stop** to end it. Logs are written to `stress_test.log` in the project directory.

## GitHub Actions

A workflow file in `.github/workflows/go-build.yml` builds the project automatically on push and pull request for Ubuntu and Windows runners.

## Disclaimer

This tool is intended for legitimate equipment testing only. Do not use it for malicious purposes.
