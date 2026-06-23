apiVersion: v1
kind: Namespace
metadata:
  name: kubesa-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubesa
  namespace: kubesa-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubesa-bootstrap-admin
subjects:
  - kind: ServiceAccount
    name: kubesa
    namespace: kubesa-system
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Secret
metadata:
  name: kubesa-backend-env
  namespace: kubesa-system
type: Opaque
stringData:
  DATABASE_URL: "__DATABASE_URL__"
  JWT_SECRET: "__JWT_SECRET__"
  DEFAULT_ADMIN_EMAIL: "__DEFAULT_ADMIN_EMAIL__"
  DEFAULT_ADMIN_PASSWORD: "__DEFAULT_ADMIN_PASSWORD__"
  DEFAULT_ADMIN_USERNAME: "__DEFAULT_ADMIN_USERNAME__"
  DEFAULT_ADMIN_NAME: "__DEFAULT_ADMIN_NAME__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubesa-backend-env
  namespace: kubesa-system
data:
  PORT: "8080"
  K8S_PROXY_URL: "http://127.0.0.1:8001"
  CORS_ALLOWED: "__CORS_ALLOWED__"
  DEFAULT_DOMAIN: "__DEFAULT_DOMAIN__"
  DEFAULT_REGISTRY_ENABLED: "true"
  DEFAULT_REGISTRY_NAME: "Default Registry"
  TCP_PROXY_HOST: "__TCP_PROXY_HOST__"
  TCP_PROXY_NAMESPACE: "kubesa-system"
  TCP_PROXY_NAME: "tcp-proxy"
  TCP_PROXY_PORT_START: "__TCP_PROXY_PORT_START__"
  TCP_PROXY_PORT_END: "__TCP_PROXY_PORT_END__"
  SERVER_IP: "__SERVER_IP__"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubesa-backend
  namespace: kubesa-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubesa-backend
  template:
    metadata:
      labels:
        app: kubesa-backend
    spec:
      serviceAccountName: kubesa
      containers:
        - name: backend
          image: "__BACKEND_IMAGE__"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
              name: http
          envFrom:
            - secretRef:
                name: kubesa-backend-env
            - configMapRef:
                name: kubesa-backend-env
        - name: kubectl-proxy
          image: bitnami/kubectl:latest
          imagePullPolicy: IfNotPresent
          args:
            - proxy
            - --address=0.0.0.0
            - --port=8001
            - --accept-hosts=.*
---
apiVersion: v1
kind: Service
metadata:
  name: kubesa-backend
  namespace: kubesa-system
spec:
  type: ClusterIP
  selector:
    app: kubesa-backend
  ports:
    - name: http
      port: 8080
      targetPort: http
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubesa-backend
  namespace: kubesa-system
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - "__BACKEND_HOST__"
      secretName: kubesa-backend-tls
  rules:
    - host: "__BACKEND_HOST__"
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kubesa-backend
                port:
                  number: 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubesa-frontend-env
  namespace: kubesa-system
data:
  NODE_ENV: "production"
  PORT: "3000"
  API_BASE_URL: "https://__BACKEND_HOST__"
  LOAD_BALANCER_IP: "__LOAD_BALANCER_IP__"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubesa-frontend
  namespace: kubesa-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubesa-frontend
  template:
    metadata:
      labels:
        app: kubesa-frontend
    spec:
      containers:
        - name: frontend
          image: "__FRONTEND_IMAGE__"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
              name: http
          envFrom:
            - configMapRef:
                name: kubesa-frontend-env
---
apiVersion: v1
kind: Service
metadata:
  name: kubesa-frontend
  namespace: kubesa-system
spec:
  type: ClusterIP
  selector:
    app: kubesa-frontend
  ports:
    - name: http
      port: 3000
      targetPort: http
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubesa-frontend
  namespace: kubesa-system
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - "__FRONTEND_HOST__"
      secretName: kubesa-frontend-tls
  rules:
    - host: "__FRONTEND_HOST__"
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kubesa-frontend
                port:
                  number: 3000
