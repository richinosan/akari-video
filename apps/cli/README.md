# AKARI CLI (`apps/cli`)

Go + Cobra front-end for headless AKARI Video. The IPC contract is
[ConnectRPC](https://connectrpc.com) (`proto/akari/v1/orchestrator.proto`).

v0 delegates to existing Node CLIs:

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

## Build

```sh
# from repo root (mise recommended)
mise install
buf generate
cd apps/cli && go build -o ../../bin/akari ./cmd/akari
```

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
