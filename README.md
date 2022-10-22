# wrkb

WRK benchmark run WRK multiple times and pick stats.
After benchmarks done it prints bench results.

```
go run cmd/wrkb/main.go main http://127.0.0.1
```

where 'main' is target local process name

```
Process "main" starts with:
cpu 0.000000
threads 2
mem 1011712
disk 245456

num|    rps|  latency|  cpu| thr|         rss
---------------------------------------------
  1|  32060|  30.25µs| 0.41|   2|     1032192
  2|  55080|  31.38µs| 0.63|   2|     1040384
  4|  90300|  36.61µs| 0.96|   2|     1056768
  8| 100070|   56.1µs| 0.99|   2|     1089536
 16| 104090|   97.3µs| 1.03|   2|     1155072
 32| 105060|  179.8µs| 1.00|   2|     1286144

Best:
  8| 100070|   56.1µs| 0.99|   2|     1089536
```
