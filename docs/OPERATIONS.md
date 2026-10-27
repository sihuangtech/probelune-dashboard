# Operations / 部署与运维

## Source deployment

On Linux or macOS, install Go 1.27.3 and Node.js 26.12.1, then run:

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
./scripts/build.sh
./dist/probelune-dashboard --listen 127.0.0.1:8080 --data ./data --web ./dist/web
```

Keep the process running. In a second terminal in the same repository, run `cat ./data/setup-token`. Visit `http://localhost:8080`, enter this token and create the first administrator. Stop with Ctrl-C; repeat the executable command to restart on the same data. Use absolute `--data` and `--web` paths when launching from another working directory.

The build produces a CGO-free server executable and the `dist/web/` assets. The running process creates its SQLite database and encryption key inside the data directory. The server reads `--data`, `--web`, `--listen`; `--backup NEW_DIRECTORY` creates a snapshot and exits.

## Linux service with systemd

On a Debian/Ubuntu host with systemd, build using the steps above or copy the matching Linux executable and web assets from your build machine. To cross-build from macOS, use `GOOS=linux GOARCH=amd64 ./scripts/build.sh`, or `arm64` for an ARM server. Run these installation commands on the Linux host from the project directory:

```sh
sudo useradd --system --user-group --home-dir /var/lib/probelune --shell /usr/sbin/nologin probelune
sudo install -m 755 dist/probelune-dashboard /usr/local/bin/probelune-dashboard
sudo install -d -m 755 /opt/probelune/web
sudo cp -R dist/web/. /opt/probelune/web/
sudo install -m 644 deploy/probelune-dashboard.service /etc/systemd/system/probelune-dashboard.service
sudoedit /etc/systemd/system/probelune-dashboard.service
```

Set the unit's `PROBELUNE_PUBLIC_URL` to the URL you actually use, for example `https://monitor.example.com`. A local-only trial can use `http://localhost:8080`. Then start the service:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now probelune-dashboard
sudo systemctl status probelune-dashboard
sudo cat /var/lib/probelune/setup-token
```

The unit creates a private `/var/lib/probelune` directory owned by its service user, restarts after failures and starts after reboot. After configuring the reverse proxy below, open your instance URL and initialize the administrator. Logs: `sudo journalctl -u probelune-dashboard -n 50`. For a program update, back up first, stop the service, replace the executable and web assets, then start the service on its existing data directory.

## Docker Compose

```sh
docker compose build --pull
docker compose up -d
docker compose exec dashboard cat /data/setup-token
```

Open `http://localhost:8080` and initialize the administrator with the token. The Dockerfiles specify verified stable patch versions; see [Dependencies](DEPENDENCIES.md). `--pull` refreshes those base images. For an update that also refreshes Alpine packages, run `docker compose build --pull --no-cache` before `docker compose up -d`.

Compose runs UID 65532 with a read-only root filesystem, writable data volume and `/tmp`. Bind-mounted directories must be writable by UID 65532. One server process owns one SQLite database; this release does not support multiple writer replicas or network-filesystem database volumes.

## HTTPS and public URL

Set `PROBELUNE_PUBLIC_URL=https://monitor.example.com` in the service/Compose environment, and set the same URL under Settings. The environment controls trusted browser Origin and Secure cookies behind an HTTPS proxy; the saved setting controls notification links. Mount a persistent data volume, keep the server behind the proxy and forward the original Host header. Serve at the origin root, not a URL subdirectory. `/healthz` is unauthenticated and reveals only health/version. Never publish the data directory as web content.

An example Caddy configuration is in `deploy/Caddyfile.example`. The systemd unit and Compose service both listen on loopback at port 8080 for this proxy.

初始化令牌保存在私有数据目录的 `setup-token`，只供首次管理员创建使用。初始化后无需保留在浏览器或配置文件。管理员创建成员，角色为 admin、operator、viewer。客户端 HTTP 仅用于明确允许的受控开发网络，正式部署使用 HTTPS。私有 CA 可以通过主机受信任根证书配置接入，不能关闭全局 TLS 校验。

## Backup and restoration

```sh
# Host binary: choose a directory that does not yet exist.
./dist/probelune-dashboard --data ./data --backup ./backup-2026-10-28
# Container: snapshot inside its writable persistent volume.
docker compose exec dashboard probelune-dashboard --data /data --backup /data/backup-2026-10-28
docker compose cp dashboard:/data/backup-2026-10-28 ./backup-2026-10-28
```

The backup includes a SQLite `VACUUM INTO` snapshot, `encryption.key` and a SHA-256 manifest. Protect all three as secrets; encrypt off-host copies. The key is required to decrypt users, sessions, monitor headers, nodes and notification credentials. Raw check evidence is not separately encrypted: protect the filesystem/volume too. Never copy only a live `probelune.db` while ignoring its WAL.

To restore: stop the destination server, use an empty private data directory, verify the two checksums against `manifest.json`, copy `probelune.db` and `encryption.key`, fix ownership/permissions, and start the **same release** against that directory. Reapply environment configuration, proxy settings and external Firebase credentials. Those external files are not in the database backup. Confirm login, monitor configuration, readable history and node heartbeats before switching traffic. Keep the original data and backup until verification completes.

Restore to the same panel URL to preserve node identity binding. For a different URL, rotate node credentials and re-register explicitly with `--register`; never silently send saved identities to another server. Missing encryption keys and unsupported schema versions fail startup. Do not generate a replacement key to try to repair an existing database.

## Upgrade / rollback

Read the changelog and contract version, create a backup, stop the old process, fetch the desired release, rebuild/pull its image and start on the existing volume. Do not run `docker compose down -v` during upgrades. The current schema is version 1. For a future incompatible migration, rollback means restoring both the prior program and its matching database/key backup. A copied database from a newer schema is not a supported downgrade route.

## Data semantics and retention

Raw samples default to 7 days (configurable 1–90); 5-minute buckets are kept 90 days, hourly buckets 365 days; incidents/audit/notification history default to 180 days (7–365 configurable); one-off diagnostic output is kept 7 days. Cleanup runs hourly. History shows availability and coverage separately. Paused/archived configuration periods are excluded from expected checks; maintenance remains included because checks continue. Missing/offline/executor-error samples reduce coverage rather than counting as success or target failure. No valid samples means no availability percentage.

节点租约 15 分钟；断连期间仅在租约有效期内继续旧任务，结果最多暂存 24 小时/100 MiB。迟到或旧配置结果只补历史，不改变新鲜实时状态。60 秒无心跳显示离线；时间偏差达到 30 秒停止分配。维护结束后重新累计确认次数。

## Operational limits

- Check interval 15–86400 seconds, timeout 1–60 seconds and strictly below the interval; no overlapping execution of the same assignment.
- One diagnostic request targets 1–20 nodes; one monitor binds 1–50. UI refresh uses polling, not a streaming subscription.
- Raw history is paginated (up to 1000 rows); aggregate history covers up to one year. Large result bodies are bounded.
- Monitor, node and member lists are intended for small instances; large-team load/SLO targets require benchmark validation.
- The panel cannot notify reliably while it is down. Configure an independent external health check of `/healthz`, disk space and backups.

Logs avoid raw credentials. A failed delivery is visible under Alerts; `sent` means provider/channel acceptance, never proof that a human read it. FCM/APNs and SMTP provider credentials must be supplied by the deployer.
