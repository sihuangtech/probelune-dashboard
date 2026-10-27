# Probelune Dashboard

[简体中文](README.zh-CN.md) · [Product and PRD](https://github.com/sihuangtech/probelune) · [Probe node](https://github.com/sihuangtech/probelune-node)

**Probelune (巡汐)** is a self-hosted service monitoring and network diagnosis system. Use the dashboard to follow checks from your own probe nodes, investigate failures across locations, and manage alerts.

## Features

- HTTP/HTTPS, TCP, ICMP and DNS monitoring with explicit node assignment, groups, tags, versioned configuration, pause and archive.
- Location matrix, availability **and** coverage, raw evidence, certificate validity, timing and 5-minute/hourly history.
- Consecutive failure/recovery confirmation; node outages remain separate from target failures.
- One-off checks including MTR/traceroute, cancellation, and redacted JSON/CSV exports.
- Incidents, acknowledgement, notes, timed silence, maintenance windows, SMTP and signed Webhooks with retries.
- Admin/operator/viewer roles, encrypted credentials, one-time node registration and mobile pairing, revocable sessions, optional public status page.
- [Desktop](https://github.com/sihuangtech/probelune-desktop) and [Flutter mobile](https://github.com/sihuangtech/probelune-mobile) clients; optional self-managed FCM delivery.

Deploy nodes on the public or private networks you want to observe. Initial releases target personal and small-team installations.

## Source deployment

On Linux or macOS, install [Go](https://go.dev/doc/install) and [Node.js](https://nodejs.org/en/download). The current verified versions are Go **1.27.3** and Node.js **26.12.1**. Build and start the dashboard:

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
./scripts/build.sh
./dist/probelune-dashboard --listen 127.0.0.1:8080 --data ./data --web ./dist/web
```

Leave that terminal running. Open a second terminal in the **same project directory** and read the initialization token:

```sh
cat ./data/setup-token
```

Open `http://localhost:8080`, enter the token, and create an administrator with a password of at least 12 characters. Stop the server with Ctrl-C. To restart, run the same executable command with the same data directory.

The build installs the locked web dependencies and creates `dist/probelune-dashboard` and `dist/web/`. The running service needs the executable, web assets and writable data directory; SQLite is created automatically. Keep `data/` when upgrading. For a persistent Linux service, follow the [host deployment](docs/OPERATIONS.md#linux-service-with-systemd) steps.

## Quick start: Docker Compose

Requires Docker with Compose and access to official image/package registries.

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
docker compose build --pull
docker compose up -d
docker compose exec dashboard cat /data/setup-token
```

Open `http://localhost:8080`, enter the one-time setup token, and create an administrator with a password of at least 12 characters. Keep the token private. Creating the administrator closes initial setup.

The Dockerfiles specify the verified stable patches: Node.js **26.12.1**, Go **1.27.3**, and Alpine **3.24.3**. See the [dependency record](docs/DEPENDENCIES.md) for sources and refresh commands.

## Configure monitoring

1. In **Settings**, set your externally reachable HTTPS instance URL and display timezone.
2. In **Nodes**, create a node and copy its ten-minute, single-use registration token.
3. Deploy [Probelune Node](https://github.com/sihuangtech/probelune-node#build-and-register). Grant private CIDRs only when intentionally monitoring those networks.
4. Create a monitor, select nodes, optionally run a one-off preflight, and wait for confirmation samples.
5. Configure and test notification channels; connect mobile/desktop clients as needed.

Both startup options bind to loopback by default. Configure an HTTPS reverse proxy for remote access using the [deployment guide](docs/OPERATIONS.md). Compose stores SQLite and encryption keys in the persistent `dashboard-data` volume; keep it when upgrading.

## Development

For live web development, run the Go server on port 8080, then run `npm ci` and `npm run dev` from `web/` in a second terminal. Vite proxies `/api` to the backend. See the [development guide](docs/ARCHITECTURE.md#developer-loop) for the full commands.

## Documentation

- [Deployment](docs/OPERATIONS.md): host service, Docker, HTTPS, backups and upgrades.
- [Development](docs/ARCHITECTURE.md): architecture and local development.
- [Notifications](docs/NOTIFICATIONS.md): SMTP, Webhooks and mobile push configuration.
- [Dependencies](docs/DEPENDENCIES.md): verified versions and update policy.
- [API](https://github.com/sihuangtech/probelune-contracts/tree/main/docs): HTTP endpoints and node protocol.
- [Acceptance](https://github.com/sihuangtech/probelune/blob/main/docs/acceptance.md): observed results and remaining deployment checks.
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
- [Security](SECURITY.md)

Checks: `cd web && npm run lint && npm run build`; `cd server && GOWORK=off go vet ./... && GOWORK=off go test ./... && GOWORK=off go build ./cmd/server`.

GPL-3.0-or-later; see [LICENSE](LICENSE). Review the validation record before relying on a particular operating system, push provider or production workload.


## Star History

[![Star history](https://api.star-history.com/svg?repos=sihuangtech/probelune-dashboard&type=Date)](https://star-history.com/#sihuangtech/probelune-dashboard&Date)

## Contact

- Developer: [Snake Konginchrist](https://github.com/Snake-Konginchrist)
- Email: [developer@skstudio.cn](mailto:developer@skstudio.cn)
- Feedback: [GitHub Issues](https://github.com/sihuangtech/probelune-dashboard/issues)
