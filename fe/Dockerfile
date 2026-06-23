# Dockerfile untuk Remix Vite App
FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./
RUN npm ci && npm cache clean --force

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Build the app
RUN npm run build

# Production image, copy all the files and run remix
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production

# Install only production dependencies for runtime
COPY package.json package-lock.json* ./
RUN npm ci --only=production && npm cache clean --force

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 remix

# Copy built application
COPY --from=builder --chown=remix:nodejs /app/build ./build
COPY --chown=remix:nodejs package.json ./package.json

USER remix

EXPOSE 3000

ENV PORT=3000

CMD ["npm", "start"]