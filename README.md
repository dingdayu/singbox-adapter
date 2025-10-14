# Go Project Template (Gin)

> 中文 [README.zh.md](README.zh.md).

An opinionated, ready-to-use Go web project template using Gin. It includes:

- Gin router and middleware (logging, recovery, rate limiting, auth)
- Full observable scheme inheritance (opentelemetry: log, trace, metric)
- Asynchronous task management (async or cron)
- CI and linting examples
- Dockerfile and image build targets
- A script/workflow to rename the module when you create a repo from this template

This README provides a concise quick start, development commands, and configuration notes.

## Quick start

1. Create a new repository from this template using GitHub's "Use this template" button.
2. Rename the module to match your repo path (see options below).
3. Run the app locally.

### Rename Go module

Option A — local script (recommended):

```bash
./scripts/rename-module.sh github.com/<your>/<repo>
```

Option B — GitHub Actions:

Open your repository Actions tab and run the "Rename Go module (one-time)" workflow. Leave the input empty to use the default `github.com/<owner>/<repo>`.

## Development

Install dependencies and run the server:

```bash
make tidy
make run
```

Open the health and example endpoints:

- http://localhost:8080/healthz
- http://localhost:8080/version

## Test & Lint

Run unit tests and linters:

```bash
make test
make lint
```

## Docker

Build the production image:

```bash
make docker-build
```

## Configuration

Set environment variables to control runtime behavior:

- PORT: listening port (default: 8080)
- APP_ENV: debug | release | test (from gin)
- APP_NAME: application name

Other configuration values are loaded from the project's config package. See `pkg/config` or `config.yaml` for details.

## Version info

Build-time version information is injected into `model/entity/version.go` using `-ldflags` in the Makefile/CI. See the Makefile and CI workflow for examples.

## Project structure (selected)

- `cmd/` — entry points and CLI commands
- `api/` — HTTP server and controllers
- `pkg/` — reusable packages (config, jwt, logger, otel)
- `model/` — data models and DAOs
- `scripts/` — helper scripts (module rename)

## Notes

- This template aims to be minimal and practical. Feel free to remove or replace components you don't need.
- If you have questions about running or customizing the template, open an issue in your fork.
