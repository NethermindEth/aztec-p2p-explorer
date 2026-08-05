# Build the frontend first, then the backend which will consume the frontend, then the runtime
FROM node:25-slim AS node-build-env
ENV P2P_BACKEND_EXPLORER_DEPENDENCIES_LAST_UPDATED=2024-11-28
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME/bin:$PATH"
RUN corepack enable

FROM node-build-env AS p2p-explorer-frontend-builder-base
COPY ./frontend ./app
WORKDIR /app

FROM p2p-explorer-frontend-builder-base AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile

FROM p2p-explorer-frontend-builder-base AS p2p-explorer-frontend-builder
ARG APP_VERSION="0.0.0"
ARG VITE_AZTEC_NETWORK="aztec-testnet"
ENV VITE_APP_VERSION=$APP_VERSION
ENV VITE_AZTEC_NETWORK=$VITE_AZTEC_NETWORK
ENV VITE_DISABLE_SYNC_STATUS=false
ENV VITE_MODE="production"
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm build

# Now we build the backend. We will copy the frontend assets from the frontend build
# so that we can embed them into the backend binary
FROM golang:1.26.5-alpine AS go-build-env

# Update the date to bust the cache and trigger downloading dependencies
ENV P2P_EXPLORER_DEPENDENCIES_LAST_UPDATED=2024-11-28
# Install build dependencies
RUN apk add --no-cache \
    file \
    make \
    git

FROM go-build-env AS p2p-explorer-backend-builder
ARG GITHUB_SHA
ENV GITHUB_SHA=$GITHUB_SHA

WORKDIR /code
COPY go.* .
RUN go mod download
COPY --from=p2p-explorer-frontend-builder /app/dist/ ./frontend/dist/
COPY . .
# Ensure static linking for Alpine compatibility
RUN CGO_ENABLED=0 make build-production

# Set up the final runtime environment based on Debian Bookworm
FROM alpine:3.24 AS aztec-p2p-explorer-runtime

# Install CA certificates
RUN apk add --no-cache ca-certificates file

# Install runtime dependencies 
# Copy the built application to the final runtime image
COPY --from=p2p-explorer-backend-builder /code/aztec-p2p-explorer /usr/local/bin/

# Copy the nebula binary
COPY --from=nethermindeth/nebula-crawler:sha-7fa14fa /usr/local/bin/nebula /usr/local/bin/
RUN chmod +x /usr/local/bin/nebula

# Copy the GeoIP database from the geoipupdate stage
COPY GeoIPDB /var/lib/GeoIP

# Set the entry point for the container to run the p2p-explorer with the GeoIP directory specified
ENTRYPOINT ["/usr/local/bin/aztec-p2p-explorer"]

# Default command (can be overridden)
CMD ["server"]
