# AKARI CLI (`apps/cli`)

**English** | [日本語](./README.ja.md)

Go + Cobra entry point for AKARI Video. The IPC contract is
[ConnectRPC](https://connectrpc.com) (`proto/akari/v1/orchestrator.proto`).

- `akari` — launch the desktop app (Electron)
- `akari cli` — headless commands; render work delegates to a bundled Node orchestrator worker

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

## Build

```sh
# from repo root (mise recommended)
mise install
mise run gen-proto
mise run build              # → dist/akari-<platform>/ (app + cli + bundle)
mise run build-linux        # → dist/akari-linux-x64/
mise run build-win64        # → dist/akari-win-x64/
```

Each distribution directory contains:

```text
dist/akari-linux-x64/
  akari                 # launcher + `akari cli` headless commands
  lib/akari-bundle/     # Node orchestrator runtime
  lib/app/              # Electron desktop app
```

`build*` runs `bundle-orchestrator`, packages the Electron shell when the host can
build for the target platform, then cross-compiles the Go entry binary.

Pinned in repo-root `mise.toml`: `node`, `ffmpeg`, `go`, `buf`, `golangci-lint`, `govulncheck` (+ `mise.lock`).

CI (`.github/workflows/go-cli.yml`): `golangci-lint` + `go test ./...` on PR/push.
Weekly `govulncheck` (`.github/workflows/go-cli-security.yml`) opens/updates a dependency PR when needed.

## Commands

```sh
akari
akari /path/to/project
akari cli version
akari cli serve --addr 127.0.0.1:7707
akari cli render plan <project-root>
akari cli render run <project-root> --approve-plan
akari cli batch --plan-only <project-a> <project-b>
akari cli batch --approve-plan <project-a> <project-b>
```

`--repo-root` or `AKARI_VIDEO_ROOT` points at the monorepo when developing without
`lib/akari-bundle/`. `AKARI_ORCHESTRATOR_ADDR` overrides the worker address
(default `127.0.0.1:7707`). `AKARI_NODE` overrides the Node binary used to spawn
the worker.

## IPC

`OrchestratorService` is implemented by the Node worker:

- bundled: `lib/akari-bundle/packages/orchestrator/bin/akari-orchestrator.mjs serve`
- dev: `node packages/orchestrator/bin/akari-orchestrator.mjs serve`

Go commands are ConnectRPC clients. Health check: `GET /healthz`.
