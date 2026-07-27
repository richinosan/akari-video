# AKARI Orchestrator (`packages/orchestrator`)

**English** | [日本語](./README.ja.md)

Node ConnectRPC worker for `OrchestratorService` (`proto/akari/v1/orchestrator.proto`).
Delegates to in-process `edit-lint` and `render-cut` (`runCli`).

## Serve

```sh
# from repo root (after npm install / mise install)
node packages/orchestrator/bin/akari-orchestrator.mjs serve --addr 127.0.0.1:7707
```

Go `akari serve` exposes the same contract on the same default port — run one worker at a time.

## Codegen

`mise run gen-proto` generates both Go (`apps/cli/gen`) and ES (`packages/orchestrator/gen`) stubs.

## Test

```sh
mise run orchestrator-test
```
