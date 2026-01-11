# --- Build stage ---
FROM node:22-alpine AS builder

WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy dependency files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source
COPY . .

# Build Nitro server output
RUN pnpm build

# --- Runtime stage ---
FROM node:22-alpine AS runner

WORKDIR /app

# Install pnpm (Nitro sometimes expects it for runtime hooks)
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy Nitro output only (standalone server)
COPY --from=builder /app/.output ./.output

# Nitro default entrypoint
CMD ["node", ".output/server/index.mjs"]

# Nitro defaults to port 3000 unless configured otherwise
EXPOSE 3000