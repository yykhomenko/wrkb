# wrkb

`wrkb` is a lightweight CLI tool for HTTP load testing. It fans out concurrent workers, measures latency with HDR histograms, and reports request/response statistics together with optional process metrics for the target service.

## Features
- 🚀 Sequential connection sweeps (e.g., `1,2,4,8…`) with total RPS limits
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
wrkb [options] <url>
```

The URL can include dynamic placeholders (see below).

## Options (with examples)

| Flag | Description | Default | Example |
| --- | --- | --- | --- |
| `-p, --proc` | Process name to monitor (optional). | — | `wrkb -p pico-http http://127.0.0.1:8082/` |
| `-c, --conns` | Comma-separated connection counts to sweep. | `1,2,4,8,16,32,64,128,256` | `wrkb -c 1,4,16 http://127.0.0.1:8082/` |
| `-t, --time` | Test duration in seconds. | `1` | `wrkb -t 10 http://127.0.0.1:8082/` |
| `-n, --requests` | Total number of requests to send (`0` = unlimited). | `0` | `wrkb -n 50000 http://127.0.0.1:8082/` |
| `--rps, --rate` | Limit total requests per second across all connections (`0` = unlimited). | `0` | `wrkb --rps 2000 http://127.0.0.1:8082/` |
| `-X, --method` | HTTP method. | `GET` | `wrkb -X POST http://127.0.0.1:8082/submit` |
| `-H, --header` | Repeatable custom header(s). | — | `wrkb -H 'Authorization: Bearer xxx' -H 'Content-Type: application/json' http://127.0.0.1:8082/` |
| `-d, --data` | Request body for write methods. | — | `wrkb -X POST -d '{"id":"123"}' http://127.0.0.1:8082/submit` |
| `-v, --verbose` | Enable verbose output. | `false` | `wrkb -v http://127.0.0.1:8082/` |
| `--best-json` | Write best benchmark result to JSON (`--best-json` = stdout, `--best-json=path` = file). | — | `wrkb --best-json=best.json http://127.0.0.1:8082/` |

## Dynamic placeholders
Use templated tokens to inject randomness before each request:

- `__RANDI64_<low>_<high>__` — random int64 within the inclusive range
- `__SEQI64_<low>_<high>__` — sequential int64 within the inclusive range (wraps)
- `__RANDHEX_<len>__` — random hex string of length `<len>`
- `__RANDSTR_letters_<len>__` — random alphabetic string
- `__RANDSTR_digits_<len>__` — random numeric string
- `__RANDSTR_lettersdigits_<len>__` — random alphanumeric string

## Examples
```bash
# Benchmark a local service by process name
wrkb -p pico-http http://127.0.0.1:8082/

# POST with custom headers across multiple connection counts
wrkb -p api -X POST \
  -d '{"id":"__RANDHEX_16__"}' \
  -H 'Content-Type: application/json' \
  http://127.0.0.1:8088/

# __RANDI64__ in URL (phone range)
wrkb -p hashes http://127.0.0.1:8082/hashes/__RANDI64_380670000001_380679999999__

# __SEQI64__ in URL (sequential id)
wrkb -p hashes http://127.0.0.1:8082/items/__SEQI64_1_100000__

# __RANDHEX__ in URL (hex identifier)
wrkb -p hashes http://127.0.0.1:8082/msisdns/__RANDHEX_32__

# __RANDSTR_letters__ in payload
wrkb -p api -X POST \
  -d '{"name":"__RANDSTR_letters_12__"}' \
  -H 'Content-Type: application/json' \
  http://127.0.0.1:8088/users

# __RANDSTR_digits__ in payload
wrkb -p api -X POST \
  -d '{"pin":"__RANDSTR_digits_6__"}' \
  -H 'Content-Type: application/json' \
  http://127.0.0.1:8088/pin

# __RANDSTR_lettersdigits__ in payload
wrkb -p api -X POST \
  -d '{"token":"__RANDSTR_lettersdigits_24__"}' \
  -H 'Content-Type: application/json' \
  http://127.0.0.1:8088/tokens
```

## Reading the output
A full run prints a table per connection count plus summary stats:

```
⚙️  Preparing benchmark: 'main' [POST] for http://127.0.0.1:8088/
   Connections: [1 2 4 8 16 32 64 128 256] | Duration: 1s | Requests: 0 | Verbose: false

⚙️  Process: main
   CPU: 40.60s | Threads: 13 | Mem: 14 MB | Disk: 460 kB


┌────┬────────┬────────────┬────────┬────────┬────────┬─────────┬─────────┬─────┬────┬────────┐
│conn│     rps│     latency│    good│     bad│     err│ body req│body resp│  cpu│ thr│     mem│
├────┼────────┼────────────┼────────┼────────┼────────┼─────────┼─────────┼─────┼────┼────────┤
│   1│   53429│    18.716µs│   46930│       0│       0│   1.1 MB│   610 kB│ 0.33│  13│   14 MB│
│   2│   82968│    24.105µs│   75318│       0│       0│   1.8 MB│   979 kB│ 0.49│  13│   14 MB│
│   4│  104857│    38.146µs│   97763│       0│       0│   2.3 MB│   1.3 MB│ 0.74│  13│   14 MB│
│   8│  122924│     65.08µs│  117985│       0│       0│   2.8 MB│   1.5 MB│ 0.97│  13│   14 MB│
│  16│  126966│   126.017µs│  123982│       0│       0│   3.0 MB│   1.6 MB│ 0.99│  13│   14 MB│
│  32│  128976│   248.107µs│  127150│       0│       0│   3.1 MB│   1.7 MB│ 0.99│  13│   14 MB│
│  64│  129672│   493.551µs│  128519│       0│       0│   3.1 MB│   1.7 MB│ 0.99│  13│   14 MB│
│ 128│  128523│   995.923µs│  127794│       0│       0│   3.1 MB│   1.7 MB│ 0.99│  13│   14 MB│
│ 256│  128173│  1.997303ms│  127849│       0│       0│   3.1 MB│   1.7 MB│ 0.99│  13│   14 MB│
└────┴────────┴────────────┴────────┴────────┴────────┴─────────┴─────────┴─────┴────┴────────┘

🚀  Best result: 8 connections | 122924 RPS | 65.08µs latency 
min=20.48µs  
p50=64.511µs 
p90=88.063µs 
p99=109.055µs 
p999=154.111µs 
max=571.391µs
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
