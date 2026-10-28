# 巡汐 Probelune 面板

[English](README.md) · [产品与 PRD](https://github.com/sihuangtech/probelune) · [探测节点](https://github.com/sihuangtech/probelune-node)

巡汐 Probelune 是自托管的多位置服务监测与网络诊断面板，用于集中查看自建节点的监测结果、管理告警，并从网页发起跨位置诊断。

## 功能

- HTTP/HTTPS、TCP、ICMP、DNS 持续监测，显式选择节点，分组、标签、搜索、复制、暂停、归档及配置版本。
- 多位置矩阵、可用率与覆盖率、错误证据、网络耗时、证书有效期、长期聚合历史。
- 连续失败/恢复阈值；节点离线单独提示，不冒充所有站点宕机。
- HTTP/DNS/TCP/ICMP/MTR/路由追踪即时诊断、取消及脱敏 JSON/CSV 导出。
- 事件时间线、确认、备注、限时静默、维护窗口、SMTP、签名 Webhook 及发送重试。
- 管理员/操作员/只读成员、一次性节点注册与手机配对、会话撤销、凭据加密、公开只读状态页。
- 接入独立桌面和移动客户端；可选配置自己的 Firebase 项目发送原生推送。

在需要观察的公网或内网部署探测节点，即可接入自己的面板。初版面向个人和小团队。

## 源码部署

适用于 Linux 和 macOS。安装 [Go](https://go.dev/doc/install) 和 [Node.js](https://nodejs.org/en/download)；当前验证版本为 Go **1.27.3**、Node.js **26.12.1**。克隆、编译并启动：

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
./scripts/build.sh
./dist/probelune-dashboard --listen 127.0.0.1:8080 --data ./data --web ./dist/web
```

保持该终端运行，在**同一项目目录**打开第二个终端，读取首次初始化令牌：

```sh
cat ./data/setup-token
```

访问 `http://localhost:8080`，输入令牌，创建密码至少 12 字符的管理员。按 Ctrl-C 停止服务；下次使用同一数据目录执行相同的启动命令即可。

编译脚本会安装锁定的网页依赖，生成 `dist/probelune-dashboard` 和 `dist/web/`。运行时只需程序、网页静态文件和可写的数据目录，SQLite 数据库会自动创建。升级时保留 `data/`。Linux 常驻运行的安装命令见 [部署文档](docs/OPERATIONS.md#linux-service-with-systemd)。

## Docker 快速启动

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
docker compose build --pull
docker compose up -d
docker compose exec dashboard cat /data/setup-token
```

访问 `http://localhost:8080`，输入初始化令牌，创建密码至少 12 字符的管理员。管理员创建成功后，首次初始化入口关闭。

Dockerfile 使用当前已核对的稳定补丁版本：Node.js **26.12.1**、Go **1.27.3**、Alpine **3.24.3**。官方来源及刷新步骤见 [依赖记录](docs/DEPENDENCIES.md)。

## 配置监测

依次配置实例 HTTPS 地址与时区、创建探测节点、部署节点、创建监测并选择执行位置，再测试通知。节点默认仅允许公网目标；内网监测需显式允许 CIDR。两个同机进程不代表两个真实地区。

两种启动方式默认都只监听主机回环端口，远程访问前配置 HTTPS 反向代理。Docker 数据在 `dashboard-data` 卷中，升级时保留该卷。

## 本地开发

开发网页时，先启动 8080 后端，再在 `web/` 执行 `npm ci` 和 `npm run dev`。Vite 将 API 请求代理到后端。完整双终端命令见 [开发文档](docs/ARCHITECTURE.md#developer-loop)。

## 文档

- [部署文档](docs/OPERATIONS.md)：宿主机服务、Docker、HTTPS、备份恢复与升级。
- [开发文档](docs/ARCHITECTURE.md)：项目架构与本地开发。
- [通知配置](docs/NOTIFICATIONS.md)：SMTP、Webhook 与手机推送。
- [依赖记录](docs/DEPENDENCIES.md)：核对过的版本与更新方式。
- [API 文档](https://github.com/sihuangtech/probelune-contracts/tree/main/docs)：HTTP 接口与节点协议。
- [验收记录](https://github.com/sihuangtech/probelune/blob/main/docs/acceptance.md)：已验证结果及环境边界。
- [更新记录](CHANGELOG.md)
- [贡献指南](CONTRIBUTING.md)
- [安全报告](SECURITY.md)

验证命令：网页 `npm run lint && npm run build`；后端在 `GOWORK=off` 下执行 `go vet ./...`、`go test ./...` 和 `go build ./cmd/server`。采用 [GPL-3.0-or-later](LICENSE) 开源。

## 星标历史

[![星标历史](https://api.star-history.com/svg?repos=sihuangtech/probelune-dashboard&type=Date)](https://star-history.com/#sihuangtech/probelune-dashboard&Date)

## 联系开发者

- 开发者: [Snake Konginchrist](https://github.com/Snake-Konginchrist)
- 邮箱: [developer@skstudio.cn](mailto:developer@skstudio.cn)
- 问题反馈: [GitHub Issues](https://github.com/sihuangtech/probelune-dashboard/issues)
