# wrkb

`wrkb` is a lightweight CLI tool for HTTP load testing. It fans out concurrent workers, measures latency with HDR histograms, and reports request/response statistics together with optional process metrics for the target service.

## Features
- 🚀 Sequential connection sweeps (e.g., `1,2,4,8…`) with per-connection RPS limits
- 📊 Rich latency breakdown (min, p50, p90, p99, p999, max) backed by HDR histograms
- 🔄 Dynamic payload/URL placeholders for randomized test data
- 🧠 Intelligent “best result” pick based on RPS vs. latency ratio
- 🖥️ Optional target-process monitoring (CPU, threads, RSS, binary size) via `-p/--proc`

## Installation
### Prerequisites
- Go **1.24+** (toolchain declared in `go.mod`)

### From source
```bash
# Build
make build
# or
go build ./cmd/wrkb

# Run without installing
go run ./cmd/wrkb --help
```

### Install into `$GOBIN`
```bash
go install ./cmd/wrkb
```

## Usage
```bash
wrkb -p <process-name> [options] <url>
```

Key options:

| Flag             | Description | Default |
|------------------| --- | --- |
| `-p, --proc`     | **Required.** Process name to monitor (e.g., `pico-http`). | — |
| `-c, --conns`    | Comma-separated connection counts to sweep. | `1,2,4,8,16,32,64,128,256` |
| `-t, --time`     | Test duration in seconds. | `1` |
| `-n, --requests` | Total number of requests to send (`0` = unlimited). | `0` |
| `-X, --method`   | HTTP method. | `GET` |
| `-d, --data`     | Request body for write methods. | — |
| `-H, --header`   | Repeatable custom header(s), e.g., `-H "Authorization: Bearer …"`. | — |
| `--rps, --rate`  | Per-connection RPS cap (`0` = unlimited). | `0` |
| `-v, --verbose`  | Print per-request details. | `false` |

The URL can include dynamic placeholders (see below).

## Dynamic placeholders
Use templated tokens to inject randomness before each request:

- `__RANDI64_<low>_<high>__` — random int64 within the inclusive range
- `__RANDHEX_<len>__` — random hex string of length `<len>`
- `__RANDSTR_letters_<len>__` — random alphabetic string
- `__RANDSTR_digits_<len>__` — random numeric string
- `__RANDSTR_lettersdigits_<len>__` — random alphanumeric string

Examples:
```bash
# Phone range
wrkb -p hashes http://127.0.0.1:8082/hashes/__RANDI64_380670000001_380679999999__

# Hex identifier
wrkb -p hashes http://127.0.0.1:8082/msisdns/__RANDHEX_32__

# Alphanumeric payload
wrkb -p pico -c=1 -rps=10 -t=1 \
  -d '{"msisdn": "__RANDI64_380670000001_380679999999__"}' \
  -H 'Content-Type: application/json' \
  http://127.0.0.1:8088/t
```

## Quick start
```bash
# Benchmark a local service by process name
wrkb -p pico-http http://127.0.0.1:8082/

# POST with custom headers across multiple connection counts
wrkb -p api -X POST -c=1,2,4,8 \
  -d '{"id":"__RANDHEX_16__"}' \
  -H 'Content-Type: application/json' \
  http://localhost:8080/items
```

## Reading the output
A full run prints a table per connection count plus summary stats:

```
┌────┬────────┬────────────┬────────┬────────┬────────┬─────────┬─────────┬─────┬────┬────────┐
│conn│     rps│     latency│    good│     bad│     err│ body req│body resp│  cpu│ thr│     mem│
├────┼────────┼────────────┼────────┼────────┼────────┼─────────┼─────────┼─────┼────┼────────┤
│   1│   59851│    16.708µs│   58958│       0│       0│   766 kB│   3.2 MB│ 0.29│   2│  3.2 MB│
│  64│  290635│   220.206µs│  290168│       0│       0│   3.8 MB│   9.1 MB│ 0.83│   2│  9.1 MB│
└────┴────────┴────────────┴────────┴────────┴────────┴─────────┴─────────┴─────┴────┴────────┘

✨ Best result: 64 connections | 290635 RPS | 220.206µs latency
min=13.3µs
p50=23.1µs
p90=31.2µs
p99=48.5µs
p999=212.0µs
max=910.4µs
```

- **rps** — responses per second during the test window.
- **latency** — mean latency; min/p50/p90/p99/p999/max follow in the footer.
- **good / bad / err** — HTTP status grouping (2xx/3xx, 4xx/5xx, transport errors).
- **body req/resp** — cumulative bytes sent/received.
- **cpu/thr/mem** — delta CPU time, thread count, and RSS of the monitored process.

## Benchmark strategy
`wrkb` executes connection counts sequentially using the same target and method. At the end, it selects a “best” configuration by balancing throughput (RPS) against observed latency using a weighted score (`RPS / log10(latency_ns)`).

## Development
- Run tests: `go test ./...`
- Format: `go fmt ./...`

## License
MIT
