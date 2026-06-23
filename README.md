# Kubesa

A self-hostable deployment platform (PaaS) that runs on k3s. Kubesa builds your
apps from Git with Kaniko, pushes to an in-cluster registry, and deploys them
behind Traefik with automatic Let's Encrypt TLS. It also provisions managed
services (Postgres, Redis, MinIO, etc.) exposed through a shared TCP proxy.

## Repository layout

This is a monorepo:

- **`/`** — Go backend (API + Kubernetes orchestration).
- **`fe/`** — Remix frontend.
- **`bootstrap/`** — plain Kubernetes manifests for the first install, before
  Kubesa can deploy itself.

## Prerequisites

Kubesa-managed Ingresses (and the bootstrap ingresses) request TLS certificates
from the `letsencrypt-prod` ClusterIssuer over Traefik, so the cluster must be
set up in this order before installing Kubesa.

### 1. A server with k3s

```sh
curl -sfL https://get.k3s.io | sh -
```

k3s ships with **Traefik** (the ingress controller) and **ServiceLB** by default,
so no separate ingress controller is needed — that is why the manifests use
`traefik.ingress.kubernetes.io/*` annotations. `kubectl` reads its config from
`/etc/rancher/k3s/k3s.yaml` on the node.

### 2. DNS

Point the domains used in `bootstrap/secrets.yml` and the bootstrap ingresses at
the node's public IP (A records), for example:

- `app.example.com` (frontend)
- `api.app.example.com` (backend)
- `proxy.app.example.com` (managed TCP proxy)

DNS must resolve **before** issuing certificates — the Let's Encrypt HTTP-01
challenge is served over those domains.

### 3. cert-manager

cert-manager is **not** bundled with k3s; install it separately:

```sh
CERT_MANAGER_VERSION=v1.17.2   # check for the latest stable release
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
kubectl wait --for=condition=Available -n cert-manager deploy --all --timeout=180s
```

### 4. Let's Encrypt ClusterIssuer

Edit the email in `bootstrap/clusterissuer.yaml`, then apply it:

```sh
# set spec.acme.email in bootstrap/clusterissuer.yaml first
kubectl apply -f bootstrap/clusterissuer.yaml
kubectl get clusterissuer letsencrypt-prod   # should report READY=True
```

Without this issuer, every Ingress certificate stays `pending` and TLS never
provisions.

## Install (bootstrap)

Build the bootstrap images on the k3s node and import them into k3s containerd.
The backend builds from the repo root; the frontend builds from `fe/`:

```sh
# backend (repo root)
docker build -t kubesa-backend:bootstrap .
docker save kubesa-backend:bootstrap | sudo k3s ctr images import -

# frontend (fe/)
docker build -t kubesa-frontend:bootstrap fe/
docker save kubesa-frontend:bootstrap | sudo k3s ctr images import -
```

Then apply the manifests:

```sh
cp bootstrap/secrets.example.yml bootstrap/secrets.yml
# edit bootstrap/secrets.yml, bootstrap/backend.yaml, and bootstrap/frontend.yaml

kubectl apply -f bootstrap/namespace.yaml
kubectl apply -f bootstrap/secrets.yml
kubectl apply -f bootstrap/database.yaml
kubectl apply -f bootstrap/backend.yaml
kubectl apply -f bootstrap/frontend.yaml
```

After the UI is reachable, create Kubesa service records for the backend and
frontend in Kubesa itself. From that point forward deployments are managed
through the Kubesa API.

## Local development

In-cluster, the backend uses its `kubesa` ServiceAccount automatically (no
config needed). For local development, set `K8S_PROXY_URL` (e.g.
`http://localhost:8001`) and run `kubectl proxy` on your machine. See
`.env.example` for the rest of the backend configuration and `fe/.env.example`
for the frontend.
