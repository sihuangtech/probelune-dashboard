# 巡汐 Probelune 面板

[English](README.md) · [产品与 PRD](https://github.com/sihuangtech/probelune) · [探测节点](https://github.com/sihuangtech/probelune-node)

自托管的多位置服务监测与网络诊断面板。本仓库将 **Vue 网页和 Go 后端一起开发、发布和部署**；不同网络中的探测节点由你自行部署。

## 已实现能力

- HTTP/HTTPS、TCP、ICMP、DNS 持续监测，显式选择节点，分组、标签、搜索、复制、暂停、归档及配置版本。
- 多位置矩阵、可用率与覆盖率、错误证据、网络耗时、证书有效期、长期聚合历史。
- 连续失败/恢复阈值；节点离线单独提示，不冒充所有站点宕机。
- HTTP/DNS/TCP/ICMP/MTR/路由追踪即时诊断、取消及脱敏 JSON/CSV 导出。
- 事件时间线、确认、备注、限时静默、维护窗口、SMTP、签名 Webhook 及发送重试。
- 管理员/操作员/只读成员、一次性节点注册与手机配对、会话撤销、凭据加密、公开只读状态页。
- 接入独立桌面和移动客户端；可选配置自己的 Firebase 项目发送原生推送。

不要求官方云账号、订阅或公共探针。初版面向个人和小团队，PRD 中的容量目标不代表实测承诺。

## Docker 快速启动

```sh
git clone https://github.com/sihuangtech/probelune-dashboard.git
cd probelune-dashboard
docker compose up -d --build
docker compose exec dashboard cat /data/setup-token
```

访问 **http://localhost:8080**，输入私有初始化令牌，创建密码至少 12 字符的管理员。初始化完成后关闭该入口，不开放公开注册。

依次配置实例 HTTPS 地址与时区、创建探测节点、部署节点、创建监测并选择执行位置，再测试通知。节点默认仅允许公网目标；内网监测需显式允许 CIDR。两个同机进程不代表两个真实地区。

默认只监听主机回环端口，远程访问前配置 HTTPS 反向代理。数据在 `dashboard-data` 卷中，升级时保留该卷。

## 本地开发

要求 Go 1.27+、Node.js 26+：

```sh
./scripts/build.sh
./dist/probelune-dashboard --listen 127.0.0.1:8080 --data ./data --web ./dist/web
```

开发网页：启动 8080 后端后，在 `web/` 执行 `npm ci && npm run dev`。Vite 将 API 请求代理到后端。网页使用官方 create-vue 的 `--tsgo` 原生 TypeScript 桥接方案；容器构建使用 Debian 以兼容原生模块，最终运行容器不需要 Node。

## 文档

- [部署、HTTPS、备份恢复与升级](docs/OPERATIONS.md)
- [架构与开发](docs/ARCHITECTURE.md)
- [通知与推送配置](docs/NOTIFICATIONS.md)
- [协议和 API](https://github.com/sihuangtech/probelune-contracts/tree/main/docs)
- [跨项目验收与真实设备边界](https://github.com/sihuangtech/probelune/blob/main/docs/acceptance.md)
- [更新记录](CHANGELOG.md)、[贡献指南](CONTRIBUTING.md)、[安全报告](SECURITY.md)

验证命令：网页 `npm run lint && npm run build`；后端 `go vet ./... && go test ./... && go build ./cmd/server`。采用 [GPL-3.0-or-later](LICENSE) 开源。
