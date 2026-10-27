# Probelune Dashboard web module

This Vue/TypeScript frontend is part of the dashboard repository and is deployed with its Go backend. See the [project README](../README.md) and [architecture](../docs/ARCHITECTURE.md).

Use Node.js 26+: `npm ci`, `npm run dev` for the Vite server, `npm run lint`, and `npm run build` for type checking and production assets. Start the Go API on `127.0.0.1:8080`; Vite proxies `/api` there. Production uses the Go server to serve `dist/`.

Vue's type checker uses the native TypeScript bridge from official create-vue's `--tsgo` integration. ESLint and Prettier are the configured source tools. No separate web repository, deployment service or application database is required.
