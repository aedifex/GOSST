# 🛰️ GOSST — Go Service Skeleton Template

[![Build](https://img.shields.io/badge/build-passing-brightgreen)](#)

Minimal, CI/CD-native Go service built specifically to demonstrate real-world delivery workflows in **Harness CI/CD**.

---

## 🚀 What is GOSST?

**GOSST** is a production-minded Go web service, containerized with Docker and designed from day one to run inside a modern CI/CD pipeline — with first-class support for [Harness](https://harness.io).

It exists to:

- Demonstrate real CI → Build → Push → Deploy workflows
- Validate cloud-native deployment patterns
- Serve as a clean foundation for microservices or API development
- Showcase Harness-native pipeline automation in a minimal, reproducible way

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

## 📁 Project Layout

```
.
├── main.go         # Minimal HTTP service
├── Dockerfile      # Slim container image
├── Makefile        # Build + helper targets
├── static/         # Optional static content
├── infra/          # IaC (optional extension)
└── .harness/       # Harness CI/CD pipeline configs
```

---

## 🧪 Local Development

```bash
go run main.go
# or
make build && ./gosst
```

Visit: http://localhost:8080  
Response: `hello from GOSST`

---

## 🐳 Container Build + Push

```bash
docker build -t <your-acr>.azurecr.io/gosst:<tag> .
docker push <your-acr>.azurecr.io/gosst:<tag>
```

Designed to mirror what the Harness pipeline performs automatically.

---

## 🔭 Extension Ideas

- Add readiness + liveness probes
- Expose `/metrics` endpoint
- Introduce API routes
- Add unit tests + coverage gates
- Implement multi-environment promotion (dev → staging → prod)
- Integrate STO or policy enforcement in pipeline

---

## 🧠 Author

Built by [Christopher Black](https://github.com/aedifex)

Intent: demonstrate how simplicity and enterprise-grade CI/CD automation can coexist — especially within Harness.

---

## 🪪 License

MIT — extend it, deploy it, break it, improve it.
