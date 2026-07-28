# AKARI CLI (`apps/cli`)

**English** | [日本語](./README.ja.md)

Go + Cobra front-end for headless AKARI Video. The IPC contract is
[ConnectRPC](https://connectrpc.com) (`proto/akari/v1/orchestrator.proto`).

v0 delegates to existing Node CLIs via a bundled Node orchestrator worker:

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

The Go binary (`akari`) is the user-facing CLI. `render` / `batch` auto-start the
Node worker and call `OrchestratorService` over ConnectRPC (h2c). `akari serve`
starts the worker in the foreground.

## Build

```sh
# from repo root (mise recommended)
mise install
mise run gen-proto
mise run cli-build              # → bin/akari (+ bin/akari-bundle/)
mise run cli-build-linux        # → bin/akari-linux-amd64
mise run cli-build-windows      # → bin/akari-windows-amd64.exe
```

`cli-build*` runs `bundle-orchestrator` first, copying orchestrator + edit-lint +
render-cut into `bin/akari-bundle/` beside the Go binary.

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

`--repo-root` or `AKARI_VIDEO_ROOT` points at the monorepo when developing without
`bin/akari-bundle/`. `AKARI_ORCHESTRATOR_ADDR` overrides the worker address
(default `127.0.0.1:7707`). `AKARI_NODE` overrides the Node binary used to spawn
the worker.

## IPC

`OrchestratorService` is implemented by the Node worker:

- bundled: `bin/akari-bundle/packages/orchestrator/bin/akari-orchestrator.mjs serve`
- dev: `node packages/orchestrator/bin/akari-orchestrator.mjs serve`

Go commands are ConnectRPC clients. Health check: `GET /healthz`.
