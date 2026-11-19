# 🛰️ GOSST — Go Service Skeleton Template

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/aedifex/GOSST/tree/master.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/gh/aedifex/GOSST/tree/master)

Minimal Go web service with a complete, multi-stage build pipeline on **CircleCI**.

---

## 🚀 What is GOSST?

**GOSST** is a production-minded Go web service, containerized with Docker and designed to run inside a modern CI pipeline. Every push is formatted, vetted, tested across several Go versions, cross-compiled for four platforms, smoke-tested and packaged — all on [CircleCI](https://circleci.com).

This is not a toy example.

It exists to:

- Demonstrate a real lint → test → build → verify → package workflow
- Show fan-out / fan-in pipelines with CircleCI matrix jobs
- Serve as a clean foundation for microservices or API development
- Keep CI logic in the `Makefile`, so any CI system can drive it

Zero external dependencies — standard library only.

---

## ⚙️ CI on CircleCI

Pipeline configuration lives in [`.circleci/config.yml`](.circleci/config.yml). Every push runs:

```
fmt ─┐   ┌─ test-go1.22 ─┐   ┌─ build-linux-amd64 ──┬─ smoke ─┐
     ├───┼─ test-go1.23 ─┼───┼─ build-linux-arm64 ──┤         ├─ package
vet ─┘   └─ test-go1.24 ─┘   ├─ build-darwin-amd64 ─┤         │
                             └─ build-darwin-arm64 ─┴─────────┘
```

| Stage | What it does |
| --- | --- |
| `fmt`, `vet` | Fail fast on unformatted code or `go vet` findings |
| `test-go*` | `make test` — race detector + coverage — once per Go version (matrix) |
| `build-*` | `make build` for each OS/arch (matrix), with build identity stamped in via `-ldflags` |
| `smoke` | Boots the linux/amd64 binary and hits `/health`, `/faux`, `/version`, `/hostz` |
| `package` | Collects every binary and writes `SHA256SUMS` |

Where to look in the CircleCI UI:

- **Tests tab** — every test and sub-test, once per Go version (JUnit via [gotestsum](https://github.com/gotestyourself/gotestsum))
- **Artifacts** — `gosst-<os>-<arch>` binaries + `SHA256SUMS` on the `package` job; coverage profiles on each test job; server log on `smoke`

The pipeline builds artifacts only — it doesn't push images or deploy. The goal: clarity over complexity. No magic. Just reproducible builds.

---

## 🛣️ Routes

| Route | Returns |
| --- | --- |
| `/health` | `200` with the build's commit SHA as the body — health check *and* "which build am I talking to" |
| `/version` | JSON build identity: commit, branch, build ID, timestamp, deployer, env, Go version |
| `/hostz` | JSON host vitals: hostname, pid, uptime, os/arch, Go version, num_cpu, gomaxprocs, goroutines |
| `/runtime` | Plain-text OS + CPU architecture |
| `/whatismyip` | The caller's IP address (honours `X-Forwarded-For` behind a proxy) |
| `/user-agent` | JSON echo of the request's User-Agent, as a string |
| `/get` | httpbin-style JSON echo: method, path, query args, headers, caller IP — any method |
| `/faux` | `CI Demo!` — the "did the pipeline actually ship" canary |
| `/` | Static content from `static/` |

Every route except `/get` is GET-only: `HEAD` works, anything else gets a `405`.

Build identity on `/version` is stamped at compile time via `-ldflags -X` (see the `build` target); `BUILD_ID` and `DEPLOY_ENV` are read from the environment once at startup and fall back to defaults.

The server sets read/write/idle timeouts and shuts down gracefully on `SIGTERM` / Ctrl-C, letting in-flight requests finish (up to 10s).

Behind a load balancer, curl `/hostz` a few times and watch `hostname` change as you're bounced between replicas.

---

## 📁 Project Layout

```
.
├── main.go               # HTTP service — stdlib only
├── main_test.go          # Handler unit tests
├── routes_test.go        # Router, static-site and smoke tests (real HTTP)
├── Dockerfile            # Slim alpine image (expects a prebuilt main_linux)
├── Makefile              # fmt / vet / lint / test / build / docker targets
├── static/               # Static content served at /
└── .circleci/config.yml  # CircleCI pipeline
```

---

## 🧪 Local Development

```bash
go run .            # listens on :8000 (override with PORT)
make test           # go test -race, runs natively on any host
make check          # fmt + vet + lint + mod + test
```

Visit: http://localhost:8000/health

`make build` cross-compiles a **linux/amd64** binary named `main_linux` by default — the one the container image expects, and it won't execute on macOS. Use `go run .` for local work, or override the target:

```bash
make build GOOS=darwin GOARCH=arm64 BINARY=dist/gosst-darwin-arm64
```

---

## 🐳 Container Build + Push

Not part of the CI pipeline — run by hand when you want an image:

```bash
make build
make docker-build
make docker-push
```

Or directly:

```bash
docker build --platform linux/amd64 -t <your-acr>.azurecr.io/gosst:<tag> .
docker push <your-acr>.azurecr.io/gosst:<tag>
```

The image copies a prebuilt `main_linux`, so run `make build` first.

---

## 🔭 Extension Ideas

- Security scanning stage — `govulncheck` on the built binaries, `gosec` on source
- Deploy stage: push to Azure Container Registry → deploy to Azure Container Apps
- Split readiness from liveness, with a runtime toggle to demo rolling deploys
- Expose `/metrics` in Prometheus text format
- `/status/{code}`, `/delay?ms=`, `/burn` — make the pipeline and platform visibly misbehave on command
- Multi-stage Dockerfile so the image builds from source with no prebuilt binary
- Multi-environment promotion (dev → staging → prod)

---

## 🧠 Author

Built by [Christopher Black](https://github.com/aedifex)

Intent: demonstrate how simplicity and enterprise-grade CI/CD automation can coexist.

---

## 🪪 License

MIT — extend it, deploy it, break it, improve it.
