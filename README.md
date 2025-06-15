# PenDeploy Simple - Railway.app Clone

A simple deployment service that clones GitHub repositories, builds Docker images with build args, and deploys to Kubernetes.

## 🎯 What It Does

1. **Clone** GitHub repository
2. **Build** Docker image with environment variables as build args
3. **Push** image to registry
4. **Process** Kubernetes manifests with environment substitution
5. **Deploy** to cluster

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- `nerdctl` (or Docker)
- `kubectl` configured
- `git` command
- Running Kubernetes cluster
- Container registry (localhost:5000)

### Installation

```bash
# Clone this repository
git clone <this-repo>
cd pendeploy-simple

# Install dependencies
go mod download

# Run the service
go run main.go
```

## 📝 API Usage

### Endpoint: `POST /create-deployment`

**Request Body:**
```json
{
  "githubUrl": "https://github.com/user/my-app",
  "env": {
    "IMAGE_REGISTRY": "localhost:5000/my-app:v1",
    "DATABASE_URL": "postgres://localhost/mydb",
    "API_KEY": "secret123",
    "NODE_ENV": "production"
  }
}
```

**Response:**
```json
{
  "status": "accepted",
  "imageName": "localhost:5000/my-app:v1",
  "message": "Deployment started, processing in background..."
}
```

## 📁 Repository Requirements

Your repository must have:

### 1. **Dockerfile with ARG declarations**
```dockerfile
FROM node:18-alpine

# Declare build arguments
ARG DATABASE_URL
ARG API_KEY
ARG NODE_ENV=production

# Set environment variables
ENV DATABASE_URL=${DATABASE_URL}
ENV API_KEY=${API_KEY}
ENV NODE_ENV=${NODE_ENV}

WORKDIR /app
COPY package.json ./
RUN npm install
COPY . .

EXPOSE 8080
CMD ["npm", "start"]
```

### 2. **kubernetes/ directory with manifests**
```
my-app/
├── Dockerfile
├── kubernetes/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml (optional)
└── ... (app files)
```

### 3. **Kubernetes manifests using ${IMAGE_REGISTRY}**

**kubernetes/deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: ${IMAGE_REGISTRY}
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          value: "${DATABASE_URL}"
        - name: API_KEY
          value: "${API_KEY}"
```

**kubernetes/service.yaml:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## 🔄 Deployment Flow

1. **Git Clone**: Repository cloned to temp directory
2. **Validation**: Check Dockerfile and kubernetes/ directory exist
3. **Docker Build**: Build with all env vars (except IMAGE_REGISTRY) as build args
4. **Docker Push**: Push to specified registry
5. **Manifest Processing**: Replace ${IMAGE_REGISTRY} and other variables in YAML files
6. **Kubernetes Apply**: Deploy processed manifests to cluster
7. **Cleanup**: Remove temporary files

## 📊 Example Request

```bash
curl -X POST http://localhost:8080/create-deployment \
  -H "Content-Type: application/json" \
  -d '{
    "githubUrl": "https://github.com/myorg/my-node-app",
    "env": {
      "IMAGE_REGISTRY": "localhost:5000/my-node-app:abc123",
      "DATABASE_URL": "postgres://db.example.com/myapp",
      "API_KEY": "super-secret-key",
      "NODE_ENV": "production",
      "PORT": "8080"
    }
  }'
```

## 🔍 Monitoring Logs

The deployment process runs in background. Monitor logs:

```bash
# Run with verbose logging
go run main.go

# Example log output:
# 🚀 Starting deployment for: my-app -> localhost:5000/my-app:v1
# 🔄 Step 1: Cloning repository...
# ✅ Git clone successful
# 🔨 Step 2: Building and pushing image...
# ✅ Build and push successful
# 🎯 Step 3: Processing and applying Kubernetes manifests...
# ✅ Kubernetes deployment successful
```

## ⚙️ Configuration

Environment variables for the service:

```bash
export PORT=8080                    # Service port (default: 8080)
export GIN_MODE=release             # Gin mode (default: release)
```

## 🐛 Troubleshooting

### Common Issues

**1. "Dockerfile not found"**
- Ensure Dockerfile is in repository root

**2. "kubernetes/ directory not found"**
- Ensure kubernetes/ folder exists with YAML files

**3. "docker build failed"**
- Check if all required ARGs are declared in Dockerfile
- Verify nerdctl/docker is installed and working

**4. "kubectl apply failed"**
- Check if kubectl is configured and cluster is accessible
- Verify Kubernetes manifests syntax

**5. "docker push failed"**
- Ensure registry is running and accessible
- Check registry authentication if required

### Debug Commands

```bash
# Test registry
curl http://localhost:5000/v2/

# Test kubectl
kubectl cluster-info

# Test nerdctl
nerdctl version

# Check logs
kubectl logs -l app=my-app
```

## 🎯 Example Repository Structure

```
my-awesome-app/
├── Dockerfile
├── package.json
├── server.js
├── kubernetes/
│   ├── deployment.yaml
│   └── service.yaml
└── README.md
```

This simple structure is all you need for deployment! 🚀