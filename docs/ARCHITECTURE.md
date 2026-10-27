# Architecture and development

```text
web/src/
  components/  reusable forms, shell, charts and evidence
  views/       route-level pages
  lib/         API, types, localization, display utilities
  styles/      presentation
server/
  cmd/server/  configuration and lifecycle
  internal/api/       authenticated routes and orchestration
  internal/model/     stored entities
  internal/store/     encrypted records, raw samples and aggregation
  internal/notify/    SMTP, signed Webhook and FCM adapters
  internal/backup/    consistent SQLite/key snapshots
deploy/ scripts/ docs/
```

The backend owns all state and authorization; clients do not infer alert transitions independently. Versioned contracts are imported from `github.com/sihuangtech/probelune-contracts v0.1.1`. Normal builds have no `replace` directive or dependency on adjacent checkouts. API prefix: `/api/v1`; node identity headers are separate from user sessions.

SQLite documents use AES-GCM with kind/ID as authenticated data. Monitor edits and version history share one transaction. Result IDs are stable per monitor, version, node and scheduled period. Repeated uploads do not add samples; old policies/versions and delayed results cannot replace current state. Leases include the exact monitor and node policy snapshot. All resolved IPs and each redirect are checked before dialing.

中英文界面使用集中词典，服务端错误保留中文说明。网页显示使用实例时区，存储使用 UTC；移动端与桌面端显示设备本地时间。`service` 在初版由 monitor 的名称/分组/标签承载，没有另外强迫用户先建服务层级。

## Developer loop

```sh
cd server
go run ./cmd/server --listen 127.0.0.1:8080 --web ../web/dist
# A second terminal:
cd web
npm ci
npm run dev
```

The first command's default data directory is local to the server working directory. Keep it ignored. Official scaffolding: `npm create vue@latest` and `go mod init`; package changes use npm/go commands. Use `npm exec --yes npm-check-updates -- -u` when deliberately upgrading ranges, then `npm install`, lint and build. Do not manually edit lockfiles. The TypeScript-native bridge replaces the incompatible compiler integration without downgrading Vue or its lint tooling.

Handwritten modules are kept small; generated package/lock files are not source-module size targets. Add focused tests for security or state transitions that can regress; use actual runtime checks for deployment/client integration. See the shared acceptance record for what has and has not been exercised.
