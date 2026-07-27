# AKARI CLI（`apps/cli`）

[English](./README.md) | **日本語**

headless な AKARI Video 向けの Go + Cobra フロントエンドです。IPC 契約は
[ConnectRPC](https://connectrpc.com)（`proto/akari/v1/orchestrator.proto`）です。

v0 は既存の Node CLI に委譲します。

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

## ビルド

```sh
# リポジトリルートから（mise 推奨）
mise install
mise run gen-proto
mise run cli-build   # → bin/akari
```

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

リポジトリ内で実行しない場合は、`--repo-root` または `AKARI_VIDEO_ROOT` でモノレポのルートを指定します。

## IPC

`OrchestratorService` の提供先:

- Go: `akari serve`
- Node: `node packages/orchestrator/bin/akari-orchestrator.mjs serve`

cobra サブコマンドは現状 Go ハンドラをプロセス内で呼び出します。リモートクライアントはどちらのワーカーにも同じ ConnectRPC 契約で接続できます。
