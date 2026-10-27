# Probelune Dashboard

[简体中文](README.zh-CN.md) · [Product and PRD](https://github.com/sihuangtech/probelune) · [Probe node](https://github.com/sihuangtech/probelune-node)

**Probelune (巡汐)** is a self-hosted service monitoring and network diagnosis system. This repository contains the Vue web console and Go backend, released and deployed together. Put your own probe nodes on the networks you need to observe.

## Features

- HTTP/HTTPS, TCP, ICMP and DNS monitoring with explicit node assignment, groups, tags, versioned configuration, pause and archive.
- Location matrix, availability **and** coverage, raw evidence, certificate validity, timing and 5-minute/hourly history.
- Consecutive failure/recovery confirmation; node outages remain separate from target failures.
- One-off checks including MTR/traceroute, cancellation, and redacted JSON/CSV exports.
- Incidents, acknowledgement, notes, timed silence, maintenance windows, SMTP and signed Webhooks with retries.
- Admin/operator/viewer roles, encrypted credentials, one-time node registration and mobile pairing, revocable sessions, optional public status page.
- [Desktop](https://github.com/sihuangtech/probelune-desktop) and [Flutter mobile](https://github.com/sihuangtech/probelune-mobile) clients; optional self-managed FCM delivery.

No public probe network, cloud account or subscription is required. Initial releases target personal and small-team installations; capacity targets in the PRD are not benchmark guarantees.

## Quick start: Docker Compose

Requires Docker with Compose and access to official image/package registries.

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
docker compose up -d --build
docker compose exec dashboard cat /data/setup-token
```

Open **http://localhost:8080**, enter the one-time setup token and create an administrator with a password of at least 12 characters. Keep the token private. The initial administrator closes setup; there is no public signup.

1. In **Settings**, set your externally reachable HTTPS instance URL and display timezone.
2. In **Nodes**, create a node and copy its ten-minute, single-use registration token.
3. Deploy [Probelune Node](https://github.com/sihuangtech/probelune-node#build-and-register). Grant private CIDRs only when intentionally monitoring those networks.
4. Create a monitor, select nodes, optionally run a one-off preflight, and wait for confirmation samples.
5. Configure and test notification channels; connect mobile/desktop clients as needed.

Compose binds to loopback by default. Put an HTTPS reverse proxy in front before remote use; see [operations](docs/OPERATIONS.md). The runtime is one non-root container with SQLite in the `dashboard-data` volume. Do not remove that volume to upgrade.

## Build without Docker

Requires **Go 1.27+ and Node.js 26+**. Dependencies are locked and the Go contracts module is fetched from its own public repository.

```sh
./scripts/build.sh
./dist/probelune-dashboard --listen 127.0.0.1:8080 --data ./data --web ./dist/web
```

For development, run the Go server on port 8080, then `cd web && npm ci && npm run dev`. Vite proxies `/api` to the backend. `web/` and `server/` are modules of this repository, not separate products.

## Documentation

- [Operations, HTTPS, backup, restoration and upgrades](docs/OPERATIONS.md)
- [Architecture and development](docs/ARCHITECTURE.md)
- [Notification and push setup](docs/NOTIFICATIONS.md)
- [API and node protocol](https://github.com/sihuangtech/probelune-contracts/tree/main/docs)
- [Cross-project acceptance record](https://github.com/sihuangtech/probelune/blob/main/docs/acceptance.md)
- [Changelog](CHANGELOG.md), [contributing](CONTRIBUTING.md), [security policy](SECURITY.md)

Checks: `cd web && npm run lint && npm run build`; `cd server && go vet ./... && go test ./... && go build ./cmd/server`. The native TypeScript bridge follows the official create-vue `--tsgo` integration; the Docker web build uses Debian for its native binary compatibility. The shipped server needs no Node runtime.

GPL-3.0-or-later; see [LICENSE](LICENSE). Review the validation record before relying on a particular operating system, push provider or production workload.
