<div align="center">
 <h1> Aztec P2P Explorer </h1>
</div>

<div align="center">

[![Pull Requests welcome](https://img.shields.io/badge/PRs-welcome-ff69b4.svg?style=flat-square)](https://github.com/NethermindEth/aztec-p2p-explorer/issues)
[![Main Build](https://github.com/NethermindEth/aztec-p2p-explorer/actions/workflows/docker.yml/badge.svg)](https://github.com/NethermindEth/aztec-p2p-explorer/actions/workflows/docker.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

</div>

P2P network monitoring for [Aztec](https://aztec.network): crawls the DHT, geolocates peers, tracks client versions and sync status, and serves it all through a REST API and a web dashboard.

Live instance: https://aztecnodes.xyz

## How it works

- A [Nebula](https://github.com/dennis-tra/nebula) crawler walks the Aztec DHT on a fixed interval and writes visit data to disk.
- The processor enriches each peer — GeoIP location (MaxMind GeoLite2), ENR and status decoding, sync validation against a reference Aztec node — and stores it in Postgres.
- A feeder polls the reference node's RPC for chain tips (`node_getChainTips` on v5+ nodes, `node_getL2Tips` before that) used in sync classification.
- An Echo REST API serves peers, analytics, and map data. The React frontend is embedded in the binary and served by the same process.

## Running it

### Prerequisites

1. Docker
2. MaxMind GeoLite2 databases (free):
   - [Sign up](https://www.maxmind.com/en/geolite2/signup) for a MaxMind account
   - Install [geoipupdate](https://github.com/maxmind/geoipupdate) and configure `GeoIP.conf` with your account ID and license key (see `GeoIP-example.conf`)
   - `geoipupdate --config-file <GeoIP.conf path> --database-directory GeoIPDB`

### Steps

1. Clone and build:

```
git clone https://github.com/NethermindEth/aztec-p2p-explorer
cd aztec-p2p-explorer
docker build \
  --build-arg APP_VERSION=$(git describe --tags --always) \
  --build-arg GITHUB_SHA=$(git rev-parse HEAD) \
  -t aztec-p2p-explorer:latest .
```

2. Start a database:

```
docker compose up test-database
```

3. Run the explorer:

```
docker run -d \
  --name aztec-p2p-explorer \
  -p 8080:8080 \
  aztec-p2p-explorer:latest \
  server \
  --db-dsn "postgres://explorer:explorer@host.docker.internal:15432/explorer?sslmode=disable" \
  --rpc-url "<your Aztec node RPC URL>" \
  --network aztec-testnet \
  --bootstrap-peers "<comma-separated ENRs for the target network>"
```

4. Open http://localhost:8080 for the dashboard, http://localhost:8080/swagger/index.html for the API docs.

## Configuration

Flags can also be set via environment variables with the `AZTEC_P2P_EXPLORER_` prefix (e.g. `AZTEC_P2P_EXPLORER_RPC_URL`). Run `server --help` for the full list.

| Flag | Default | Purpose |
|---|---|---|
| `--rpc-url` | — (required) | Aztec node RPC used as the sync reference |
| `--db-dsn` | localhost dev DSN | Postgres connection string |
| `--network` | `aztec-testnet` | Network to crawl (`aztec-testnet`, `aztec-mainnet`) |
| `--bootstrap-peers` | — | ENRs the crawler starts from |
| `--interval` | `30m` | Crawl interval |
| `--disable-crawler` | `false` | Serve the API without crawling |
| `--maxmind-dir` | `/var/lib/GeoIP` | GeoLite2 database directory |

Environment-only settings:

- `PRIVATE_API_USERNAME` / `PRIVATE_API_PASSWORD` — BasicAuth for `/api/private`; the private API rejects all requests when unset
- `TURNSTILE_*` — optional Cloudflare Turnstile bot protection (see `.env.example`)

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md). In short: `make deps`, `docker compose up test-database`, `make test`, `make lint`. Frontend lives in `frontend/` (Vite + React, pnpm).

## Security

See [SECURITY.md](SECURITY.md).

## License

[Apache 2.0](LICENSE)
