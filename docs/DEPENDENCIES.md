# Dependency record / 依赖版本记录

Verified on **2026-10-28** against the upstream release feeds and Docker Official Images.

| Component | Docker image | Official source |
| --- | --- | --- |
| Node.js web builder | `node:26.12.1-trixie-slim` | [Node releases](https://nodejs.org/dist/index.json), [Docker tags](https://hub.docker.com/_/node) |
| npm build tool | `npm@12.2.1` | [npm registry](https://registry.npmjs.org/npm/latest) |
| Go server builder | `golang:1.27.3-alpine3.24` | [Go downloads](https://go.dev/dl/), [Docker tags](https://hub.docker.com/_/golang) |
| Runtime OS | `alpine:3.24.3` | [Alpine releases](https://www.alpinelinux.org/releases/), [Docker tags](https://hub.docker.com/_/alpine) |

Node uses Debian Trixie in the build stage to support the native TypeScript bridge. The final image runs the compiled Go program and web assets on Alpine. SQLite is provided by the Go module `modernc.org/sqlite`; no database container is needed.

Version numbers identify the verified stable patches. Docker's `--pull` option checks for a newer build of those image tags; it does not change the selected patch version. Each repository configures a weekly Dependabot check for its Dockerfile. npm is explicitly upgraded with the official `npm install --global npm@12.2.1` command during the web build; review its registry version when refreshing the toolchain.

## Refresh a Docker deployment

After reviewing upstream changes and updating image versions, rebuild with fresh base images and package indexes:

```sh
docker compose build --pull --no-cache
docker compose up -d
```

`apk add --no-cache ca-certificates tzdata` installs the versions available in the selected stable Alpine branch. Reusing a Docker build layer can reuse older installed packages; `--no-cache` forces that installation to run again. The probe image follows the same policy for MTR, traceroute and libcap. Exact installed versions can be inspected with `docker compose exec dashboard apk info -v`.

## Application dependencies

The web `package-lock.json` records npm dependencies; `go.mod` and `go.sum` record Go modules. The version audit used the official package commands:

```sh
cd web
npm exec --yes npm-check-updates -- --target latest
npm audit
cd ../server
GOWORK=off go get -u ./...
GOWORK=off go mod tidy
GOWORK=off go list -m -u all
```

The npm direct dependencies already matched the latest published versions. The lockfile was regenerated with `npm install --package-lock-only --ignore-scripts` against the official registry, and the project's `.npmrc` selects that same registry. This avoids mixed mirror origins being rejected by npm 12's remote-package policy. The Go direct dependencies were current; outdated production transitive dependencies were upgraded with `go get -u ./...` and the module files regenerated with `go mod tidy`. Both Go modules select toolchain `go1.27.3`. Registry and OS repositories change over time; this dated record describes what was checked, and each future update must pass the builds again.

官方构建刷新参数见 [Docker 文档](https://docs.docker.com/reference/cli/docker/compose/build/)。本文件记录核对日期、明确的镜像版本与更新方法；`--pull` 刷新镜像，`--no-cache` 重新执行安装步骤，二者作用不同。
