# Kubesa Bootstrap

This directory contains plain Kubernetes manifests for the first Kubesa install,
before Kubesa can deploy itself.

## Files

- `namespace.yaml`: namespace, service account, and bootstrap RBAC.
- `secrets.example.yml`: copy this to `secrets.yml`, fill real secrets, then apply it.
- `database.yaml`: bootstrap Postgres storage, deployment, and service.
- `backend.yaml`: backend config, deployment, service, and ingress.
- `frontend.yaml`: frontend config, deployment, service, and ingress.

## Install

Build the bootstrap images on the k3s node and import them into k3s containerd:

```sh
# backend repo
docker build -t kubesa-backend:bootstrap .
docker save kubesa-backend:bootstrap | sudo k3s ctr images import -

# frontend repo
docker build -t kubesa-frontend:bootstrap .
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

The backend pod runs a `kubectl proxy` sidecar because the current backend
Kubernetes client reads `K8S_PROXY_URL`. The sidecar uses the `kubesa`
ServiceAccount bound to `cluster-admin` for bootstrap simplicity.

After the UI is reachable, create Kubesa service records for the backend and
frontend in Kubesa itself. From that point forward, the GitHub workflows can
deploy through the Kubesa API.
