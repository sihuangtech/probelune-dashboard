FROM node:26.12.1-trixie-slim AS web
RUN npm install --global npm@12.2.1
WORKDIR /src/web
COPY web/package*.json ./
COPY web/.npmrc ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.27.3-alpine3.24 AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/probelune-dashboard ./cmd/server

FROM alpine:3.24.3
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 65532 probelune && adduser -S -u 65532 -G probelune probelune \
    && mkdir /data && chown 65532:65532 /data
COPY --from=server /out/probelune-dashboard /usr/local/bin/probelune-dashboard
COPY --from=web /src/web/dist /app/web
USER 65532:65532
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["probelune-dashboard"]
CMD ["--listen", "0.0.0.0:8080", "--data", "/data", "--web", "/app/web"]
