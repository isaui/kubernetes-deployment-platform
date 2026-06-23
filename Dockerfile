FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

ARG DATABASE_URL
ARG JWT_SECRET
ARG K8S_PROXY_URL
ARG CORS_ALLOWED
ARG PORT

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=1 go build -o pendeploy-handal .

# Environment variables
ENV PORT=${PORT}
ENV CORS_ALLOWED=${CORS_ALLOWED}
ENV JWT_SECRET=${JWT_SECRET}
ENV K8S_PROXY_URL=${K8S_PROXY_URL}
ENV DATABASE_URL=${DATABASE_URL}

EXPOSE ${PORT}

CMD ["./pendeploy-handal"]
