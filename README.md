# 🛰️ GOSST — Go Service Skeleton Template

[![Build](https://img.shields.io/badge/build-passing-brightgreen)](#)

Minimal, CI/CD-native Go service built specifically to demonstrate real-world delivery workflows in **Harness CI/CD**.

---

## 🚀 What is GOSST?

**GOSST** is a production-minded Go web service, containerized with Docker and designed from day one to run inside a modern CI/CD pipeline — with first-class support for [Harness](https://harness.io).

This is not a toy example.

It exists to:

- Demonstrate real CI → Build → Push → Deploy workflows
- Validate cloud-native deployment patterns
- Serve as a clean foundation for microservices or API development
- Showcase Harness-native pipeline automation in a minimal, reproducible way

Zero external dependencies — standard library only.

---

## ⚙️ Harness-First by Design

GOSST plugs directly into **Harness CI/CD**.

Typical enterprise flow:

1. Git push triggers Harness pipeline
2. Harness builds the container image
3. Image is pushed to Azure Container Registry
4. Harness deploys to Azure Container Apps

Pipeline configuration lives in `.harness/` and is structured for clarity and extensibility.

The goal: clarity over complexity. No magic. Just reproducible delivery.

---

## 🛣️ Routes

| Route | Returns |
| --- | --- |
| `/health` | `200` with the build's commit SHA as the body — health check *and* "which build am I talking to" |
| `/version` | JSON build identity: commit, branch, build ID, timestamp, deployer, env |
| `/hostz` | JSON host vitals: hostname, pid, uptime, os/arch, num_cpu, gomaxprocs, goroutines |
| `/runtime` | Plain-text OS + CPU architecture |
| `/whatismyip` | The caller's IP address |
| `/user-agent` | JSON echo of the request's User-Agent |
| `/get` | JSON echo of the request method |
| `/faux` | `CI Demo!` — the "did the pipeline actually ship" canary |
| `/` | Static content from `static/` |

Build identity on `/version` is stamped at compile time via `-ldflags -X` (see the `build` target); `BUILD_ID` and `DEPLOY_ENV` are read from the environment at runtime and fall back to defaults.

Behind a load balancer, curl `/hostz` a few times and watch `hostname` change as you're bounced between replicas.

---

## 📁 Project Layout

```
.
├── main.go         # HTTP service — stdlib only
├── main_test.go    # Handler + smoke tests
├── Dockerfile      # Slim alpine image (expects a prebuilt main_linux)
├── Makefile        # fmt / vet / lint / test / build / docker targets
├── static/         # Static content served at /
└── .harness/       # Harness CI/CD pipeline configs
```

---

## 🧪 Local Development

```bash
go run .            # listens on :8000 (override with PORT)
make test           # go test -race, runs natively on any host
make check          # fmt + vet + lint + mod + test
```

Visit: http://localhost:8000/health

Note that `make build` cross-compiles a **linux/amd64** binary named `main_linux` for the container image — it won't execute on macOS. Use `go run .` for local work.

---

## 🐳 Container Build + Push

```bash
make docker-build
make docker-push
```

Or by hand:

```bash
docker build --platform linux/amd64 -t <your-acr>.azurecr.io/gosst:<tag> .
docker push <your-acr>.azurecr.io/gosst:<tag>
```

The image copies a prebuilt `main_linux`, so run `make build` first. Designed to mirror what the Harness pipeline performs automatically.

---

## 🔭 Extension Ideas

- Split readiness from liveness, with a runtime toggle to demo rolling deploys
- Expose `/metrics` in Prometheus text format
- `/status/{code}`, `/delay?ms=`, `/burn` — make the pipeline and platform visibly misbehave on command
- Multi-stage Dockerfile so the image builds from source with no prebuilt binary
- Multi-environment promotion (dev → staging → prod)
- Integrate STO or policy enforcement in pipeline

---

## 🧠 Author

Built by [Christopher Black](https://github.com/aedifex)

Intent: demonstrate how simplicity and enterprise-grade CI/CD automation can coexist — especially within Harness.

---

## 🪪 License

MIT — extend it, deploy it, break it, improve it.
