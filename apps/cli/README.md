# AKARI CLI (`apps/cli`)

**English** | [日本語](./README.ja.md)

Go + Cobra front-end for headless AKARI Video. The IPC contract is
[ConnectRPC](https://connectrpc.com) (`proto/akari/v1/orchestrator.proto`).

v0 delegates to existing Node CLIs:

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

## Build

```sh
# from repo root (mise recommended)
mise install
mise run gen-proto
mise run cli-build   # → bin/akari
```

Pinned in repo-root `mise.toml`: `node`, `ffmpeg`, `go`, `buf`, `golangci-lint`, `govulncheck` (+ `mise.lock`).

CI (`.github/workflows/go-cli.yml`): `golangci-lint` + `go test ./...` on PR/push.
Weekly `govulncheck` (`.github/workflows/go-cli-security.yml`) opens/updates a dependency PR when needed.

## Commands

```sh
akari version
akari serve --addr 127.0.0.1:7707
akari render plan <project-root>
akari render run <project-root> --approve-plan
akari batch --plan-only <project-a> <project-b>
akari batch --approve-plan <project-a> <project-b>
```

`--repo-root` or `AKARI_VIDEO_ROOT` points at the monorepo when not run from inside it.

## IPC

`OrchestratorService` is served by `akari serve`. Cobra subcommands call the same handler
in-process today; remote workers can implement the same contract later without changing
the protobuf surface.
