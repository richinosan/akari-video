# AKARI CLI（`apps/cli`）

[English](./README.md) | **日本語**

headless な AKARI Video 向けの Go + Cobra フロントエンドです。IPC 契約は
[ConnectRPC](https://connectrpc.com)（`proto/akari/v1/orchestrator.proto`）です。

v0 は bundled な Node orchestrator ワーカー経由で既存の Node CLI に委譲します。

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

Go バイナリ（`akari`）がユーザー向け CLI です。`render` / `batch` は Node ワーカーを
自動起動し、ConnectRPC（h2c）で `OrchestratorService` を呼びます。`akari serve` は
ワーカーをフォアグラウンドで起動します。

## ビルド

```sh
# リポジトリルートから（mise 推奨）
mise install
mise run gen-proto
mise run cli-build              # → bin/akari（+ bin/akari-bundle/）
mise run cli-build-linux        # → bin/akari-linux-amd64
mise run cli-build-windows      # → bin/akari-windows-amd64.exe
```

`cli-build*` は先に `bundle-orchestrator` を実行し、orchestrator + edit-lint +
render-cut を Go バイナリ横の `bin/akari-bundle/` に同梱します。

リポジトリルートの `mise.toml` で pin: `node` / `ffmpeg` / `go` / `buf` / `golangci-lint` / `govulncheck`（`mise.lock` 同梱）。

CI（`.github/workflows/go-cli.yml`）: PR / push で `golangci-lint` + `go test ./...`。
週次の `govulncheck`（`.github/workflows/go-cli-security.yml`）は、検出時に依存更新 PR を作成または更新します。

## コマンド

```sh
akari version
akari serve --addr 127.0.0.1:7707
akari render plan <project-root>
akari render run <project-root> --approve-plan
akari batch --plan-only <project-a> <project-b>
akari batch --approve-plan <project-a> <project-b>
```

`bin/akari-bundle/` なしで開発する場合は `--repo-root` または `AKARI_VIDEO_ROOT` で
モノレポを指定します。`AKARI_ORCHESTRATOR_ADDR` でワーカーアドレス（既定 `127.0.0.1:7707`）、
`AKARI_NODE` でワーカー起動に使う Node を上書きできます。

## IPC

`OrchestratorService` の実装は Node ワーカーです。

- 同梱: `bin/akari-bundle/packages/orchestrator/bin/akari-orchestrator.mjs serve`
- 開発: `node packages/orchestrator/bin/akari-orchestrator.mjs serve`

Go コマンドは ConnectRPC クライアントです。ヘルスチェック: `GET /healthz`。
