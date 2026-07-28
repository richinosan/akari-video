# AKARI CLI（`apps/cli`）

[English](./README.md) | **日本語**

AKARI Video の Go + Cobra 入口です。IPC 契約は
[ConnectRPC](https://connectrpc.com)（`proto/akari/v1/orchestrator.proto`）です。

- `akari` — デスクトップアプリ（Electron）を起動
- `akari cli` — ヘッドレスコマンド（レンダー処理は bundled Node ワーカーへ委譲）

- `packages/edit-lint/bin/edit-lint.mjs`
- `packages/render-cut/bin/render-cut.mjs`

## ビルド

```sh
# リポジトリルートから（mise 推奨）
mise install
mise run gen-proto
mise run build              # → dist/akari-<platform>/（app + cli + bundle）
mise run build-linux        # → dist/akari-linux-x64/
mise run build-win64        # → dist/akari-win-x64/
```

配布ディレクトリの構成:

```text
dist/akari-linux-x64/
  akari                 # ランチャー + `akari cli` ヘッドレスコマンド
  lib/akari-bundle/     # Node orchestrator ランタイム
  lib/app/              # Electron デスクトップアプリ
```

`build*` は `bundle-orchestrator` のあと、ホストが対象プラットフォーム向けに
Electron をパッケージし、Go 入口バイナリをクロスコンパイルします。

リポジトリルートの `mise.toml` で pin: `node` / `ffmpeg` / `go` / `buf` / `golangci-lint` / `govulncheck`（`mise.lock` 同梱）。

CI（`.github/workflows/go-cli.yml`）: PR / push で `golangci-lint` + `go test ./...`。
週次の `govulncheck`（`.github/workflows/go-cli-security.yml`）は、検出時に依存更新 PR を作成または更新します。

## コマンド

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

`lib/akari-bundle/` なしで開発する場合は `--repo-root` または `AKARI_VIDEO_ROOT` で
モノレポを指定します。`AKARI_ORCHESTRATOR_ADDR` でワーカーアドレス（既定 `127.0.0.1:7707`）、
`AKARI_NODE` でワーカー起動に使う Node を上書きできます。

## IPC

`OrchestratorService` の実装は Node ワーカーです。

- 同梱: `lib/akari-bundle/packages/orchestrator/bin/akari-orchestrator.mjs serve`
- 開発: `node packages/orchestrator/bin/akari-orchestrator.mjs serve`

`akari cli` は ConnectRPC クライアントです。ヘルスチェック: `GET /healthz`。
