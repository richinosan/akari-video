# AKARI Orchestrator（`packages/orchestrator`）

[English](./README.md) | **日本語**

`OrchestratorService`（`proto/akari/v1/orchestrator.proto`）の Node ConnectRPC ワーカーです。
`edit-lint` / `render-cut` の `runCli` をプロセス内で委譲します。

## 起動

```sh
# リポジトリルートから（npm install / mise install 後）
node packages/orchestrator/bin/akari-orchestrator.mjs serve --addr 127.0.0.1:7707
```

Go の `akari serve` と同じ契約・同じ既定ポートです。同時に両方は起動しないでください。

## コード生成

`mise run gen-proto` で Go（`apps/cli/gen`）と ES（`packages/orchestrator/gen`）のスタブを生成します。

## テスト

```sh
mise run orchestrator-test
```
