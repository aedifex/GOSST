# 🛰️ GOSST — Go Service Skeleton Template

[![Build](https://img.shields.io/badge/build-passing-brightgreen)](#)
Minimal, CI/CD-native Go service. Built for clarity, speed, and Harness-first deployment workflows.

---

## 🚀 What is GOSST?

**GOSST** is a skeletal web service written in Go, containerized with Docker, and purpose-built to integrate with modern CI/CD platforms — especially [Harness](https://harness.io).

It exists to:

- Demo real-world DevOps patterns
- Validate infra pipelines (build → push → deploy)
- Give you a clean slate to expand into microservices or API development

---

## 📁 Project Layout

```
.
├── main.go         # Basic HTTP server
├── Dockerfile      # Slim image for container runtime
├── Makefile        # Helper targets for build/push
├── static/         # Optional static file support
├── infra/          # IaC (if used)
└── .harness/       # Harness pipeline config
```

---

## 🧪 Local Development

```bash
go run main.go
# or
make build && ./gosst
```

Visit: http://localhost:8080
You should see: `hello from GOSST`

---

## 🐳 Container Build + Push

```bash
docker build -t <your-acr>.azurecr.io/gosst:<tag> .
docker push <your-acr>.azurecr.io/gosst:<tag>
```

---

## ⚙️ CI/CD (Harness-native)

GOSST is designed to slot directly into a [Harness CI/CD pipeline](https://harness.io/docs).

Typical flow:

1. **Git push** triggers pipeline
2. **Harness builds the image**
3. **Harness pushes to Azure Container Registry**
4. **Harness deploys to Azure Container Apps**

You can find pipeline configs in `.harness/`.

---

## 🔭 Future Ideas

- Add health/liveness probes
- Expose an internal `/metrics` endpoint
- Wire in a real API layer
- Add unit tests and coverage to the pipeline
- Deploy across multiple environments (dev, staging, prod)

---

## 🧠 Author

Built by [Christopher Black](https://github.com/aedifex)
Project goal: combine simplicity with industrial-grade deployment readiness.

---

## 🪪 License

MIT — use it, break it, extend it. All good.
