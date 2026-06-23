# Kubesa Bootstrap

This directory is the first-install path for Kubesa. Use it before Kubesa is
able to deploy itself.

## Required environment

```sh
export BACKEND_IMAGE=ghcr.io/your-org/kubesa-backend:latest
export FRONTEND_IMAGE=ghcr.io/your-org/kubesa-frontend:latest
export DATABASE_URL='postgresql://user:password@host:5432/db?sslmode=disable'
export JWT_SECRET='change-me'
export DEFAULT_DOMAIN=app.example.com
export DEFAULT_ADMIN_EMAIL=admin@example.com
export DEFAULT_ADMIN_PASSWORD='change-me'
export BACKEND_HOST=api.app.example.com
export FRONTEND_HOST=kubesa.app.example.com
export CORS_ALLOWED=https://kubesa.app.example.com
export LOAD_BALANCER_IP=1.2.3.4
export SERVER_IP=1.2.3.4
```

Optional:

```sh
export DEFAULT_ADMIN_USERNAME=admin
export DEFAULT_ADMIN_NAME='Default Admin'
export TCP_PROXY_HOST=proxy.app.example.com
export TCP_PROXY_PORT_START=24000
export TCP_PROXY_PORT_END=24999
```

## Install

```sh
./bootstrap/render-bootstrap.sh > /tmp/kubesa-bootstrap.yaml
kubectl apply -f /tmp/kubesa-bootstrap.yaml
```

PowerShell:

```powershell
.\bootstrap\render-bootstrap.ps1 | kubectl apply -f -
```

The backend pod runs a `kubectl proxy` sidecar because the current backend
Kubernetes client reads `K8S_PROXY_URL`. The sidecar uses the `kubesa`
ServiceAccount bound to `cluster-admin` for bootstrap simplicity.

After the UI is reachable, create Kubesa service records for the backend and
frontend in Kubesa itself. From that point forward, the GitHub workflows can
deploy through the Kubesa API.
