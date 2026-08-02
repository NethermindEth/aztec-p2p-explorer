# Contributing

## Setup

- Go 1.24+, Docker, and pnpm (for the frontend)
- `make deps` installs the dev tooling (golangci-lint, swag, sqlboiler, gosec)
- `docker compose up test-database` starts the Postgres instance the tests expect on `localhost:15432`

## Day to day

- `make test` — full suite; repo-layer tests run against the test database
- `make lint` — golangci-lint with the production build tag (same as CI)
- `make docs` — regenerate swagger docs after changing API annotations
- `make models` — regenerate sqlboiler models after a schema migration
- Frontend: `cd frontend && pnpm install && pnpm dev` (dev server proxies `/api` to `localhost:8080`)

## Pull requests

- Branch from `main` and keep PRs focused on one change
- Subjects follow conventional commits (`fix:`, `feat:`, `chore:`) — see `git log` for the house style
- CI (tests, lint, Docker build) must pass; PRs are squash-merged
