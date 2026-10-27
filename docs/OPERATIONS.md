# Operations / 部署与运维

## Deployment

Use `docker compose up -d --build`, or `scripts/build.sh` followed by the server executable. The server reads `--data`, `--web`, `--listen`; `--backup NEW_DIRECTORY` creates a snapshot and exits. Example service and Caddy configuration are in `deploy/`.

Set `PROBELUNE_PUBLIC_URL=https://monitor.example.com` in the service/Compose environment, and set the same URL under Settings. The environment controls trusted browser Origin and Secure cookies behind an HTTPS proxy; the saved setting controls notification links. Mount a persistent data volume, keep the server behind the proxy and forward the original Host header. Serve at the origin root, not a URL subdirectory. `/healthz` is unauthenticated and reveals only health/version. Never publish the data directory as web content.

Compose runs UID 65532 with a read-only root filesystem, writable data volume and `/tmp`. Bind-mounted directories must be writable by UID 65532. One server process owns one SQLite database; this release does not support multiple writer replicas or network-filesystem database volumes.

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
