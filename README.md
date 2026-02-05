Todo:
Client:
1) Validation
2) File extension allowance (whitelist)
3) Authen Author
4) Size limit

Network:
1) Restrict Acccess
2) Separate environment
3) Zero trust

Server:
1) CORS - [x]
2) temp folder - [x]
3) scan process : Check Malwatch or LMD - [x]
4) size limit
5) renamed - [x]
6) binary 
7) Rate limit - [x]

Log:
Monitor
Alert a big file transfer

### Example docker
                # Stage 1: Build the application
                FROM node:22-alpine AS build
                WORKDIR /app

                # Install a specific npm version (e.g., latest or 10.x.x)
                RUN npm install -g npm@latest

                # Copy only dependency files first to leverage build cache
                COPY package.json package-lock.json* ./
                RUN npm ci

                # Copy source and build
                COPY . .
                RUN npm run build

                # Stage 2: Production Runtime
                FROM node:22-slim AS runner
                WORKDIR /app

                # Set production environment
                ENV NODE_ENV=production

                # Copy built application output from the build stage
                # TanStack Start / Nitro typically outputs to .output
                COPY --from=build /app/.output ./.output
                COPY --from=build /app/package.json ./package.json

                # Expose the default TanStack Start port
                EXPOSE 3000

                # Start via the Nitro-generated server index
                # Direct 'node' execution is preferred for signal handling
                CMD ["node", ".output/server/index.mjs"]
