# Contributing to MicroGit

Thanks for your interest in improving MicroGit! It's a small, educational
project, so contributions of all sizes are welcome.

## Development setup

You'll need [Go](https://go.dev/dl/) 1.21 or newer.

```bash
git clone https://github.com/your-username/microgit.git
cd microgit
make build      # builds ./microgit
```

## Common tasks

| Command        | What it does                              |
| -------------- | ----------------------------------------- |
| `make build`   | Build the `microgit` binary               |
| `make test`    | Run all tests (unit + end-to-end)         |
| `make race`    | Run tests with the race detector          |
| `make cover`   | Show test coverage per package            |
| `make lint`    | `gofmt`, `go vet`, and `staticcheck`      |

## Project layout

```
main.go        Entry point; calls cmd.Execute()
cmd/           One file per command + repo.go (shared repository helpers)
utils/         Object storage (sharded, zlib-compressed) and shared helpers
e2e/           End-to-end tests that drive the compiled binary
```

## Guidelines

- Run `make lint` and `make test` before opening a pull request.
- Add or update tests for any behavior you change. Command behavior is covered
  by unit tests in `cmd/` and black-box tests in `e2e/`.
- Keep commits focused and write clear commit messages.
- CI runs the build, vet, tests (across Go versions and OSes) and lint on every
  pull request — see `.github/workflows/`.

## Reporting issues

Please include the MicroGit version (`microgit --version`), your OS, and the
steps to reproduce.
