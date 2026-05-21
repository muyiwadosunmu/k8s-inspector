# k8s-inspector

Small service that exposes Kubernetes information via HTTP endpoints.

## Quick overview
- Endpoints implemented:
  - `GET /healthz` — returns service health and build metadata
  - `GET /summary` — cluster summary (namespace count, running pods, nodes)
  - `GET /pods?namespace=` — lists pods in a namespace and aggregates container resource requests

## Run locally

Prerequisites: Go 1.26, Docker (optional), access to a Kubernetes cluster (kind/minikube or kubeconfig).

1. Run against your kubeconfig (out-of-cluster):

```bash
export KUBECONFIG=$HOME/.kube/config
go run ./cmd/inspector
```

2. Build and run the Docker image (recommended: tag with commit SHA):

```bash
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)
docker build --build-arg BUILD_TIME="$BUILD_TIME" --build-arg GIT_COMMIT="$GIT_COMMIT" -t registry.example.com/inspector:$GIT_COMMIT .
docker run --rm -p 3000:3000 registry.example.com/inspector:$GIT_COMMIT
```

## Kubernetes deployment

Apply manifests (replace image placeholder with your built image tag):

```bash
# update deployment image (recommended) and apply
kubectl set image deployment/inspector inspector=registry.example.com/inspector:$GIT_COMMIT --record
kubectl apply -f k8s/
kubectl rollout status deployment/inspector
```

Verify endpoints (from cluster or port-forward):

```bash
kubectl port-forward deploy/inspector 3000:3000 &
curl http://localhost:3000/healthz
curl http://localhost:3000/summary
curl "http://localhost:3000/pods?namespace=default"
```

## Project structure

- `cmd/inspector` — application entrypoint and HTTP handlers
- `internal/k8s` — Kubernetes client helper that supports in-cluster and kubeconfig
- `internal/pkg/web` — lightweight web helpers (middleware, error handling)
- `k8s/` — Kubernetes manifests (Deployment, Service, RBAC)
- `Dockerfile` — multi-stage build embedding `BUILD_TIME` and `GIT_COMMIT`

Organization rationale: keep public API in `cmd/inspector`, reusable infra code in `internal/`, and manifests in `k8s/`. This keeps the repo small and easy to review.

## Assumptions
- Service will read cluster objects (pods, namespaces, nodes) only — no mutating operations required.
- Users will deploy with an immutable image tag (commit SHA) — manifests contain a placeholder.

## What I'd improve with more time
- Add unit tests using `k8s.io/client-go/kubernetes/fake` for handlers.
- Add GitHub Actions workflow to run `gofmt`, `go vet`, unit tests, build image, push to registry, and update manifests.
- Add Prometheus metrics (request latencies, k8s API error counts) and OpenTelemetry traces.
- Use informers/caches to reduce API server pressure and improve resilience.
- Provide a Helm chart and more robust RBAC scoping (namespace-scoped Role + RoleBinding when possible).
