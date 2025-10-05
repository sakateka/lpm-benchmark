## LPM Benchmark Suite

This repository contains benchmarks and tests for multiple Longest Prefix Match (LPM) data structures. It focuses on measuring insertion, lookup, memory usage, and concurrency behavior across different implementations and input scales (including 1M prefixes).

### Implementations Covered
- Map-based trie ([`generic MapTrie`](https://github.com/yanet-platform/yanet2/blob/main/modules/route/internal/rib/map_trie.go))
- Patricia trie (via `github.com/kentik/patricia`)
- External `lpm` library (via `github.com/sakateka/lpm`)

Provenance note: the `MapTrie` tree here is a copy-paste from:
`https://github.com/yanet-platform/yanet2/blob/main/modules/route/internal/rib/map_trie.go`.

### What These Benchmarks Show (and Don’t)
- Benchmark results are workload- and implementation-dependent. A faster tree in one scenario is not universally “better,” and a slower tree is not universally “worse.”
- Each structure is tailored for different tradeoffs: insertion vs lookup speed, memory footprint, IPv4/IPv6 behavior, update patterns, and concurrency.
- Use these results as guidance for your target workload characteristics, not as absolute rankings.

### Benchmark Coverage
- Insertion micro-benchmarks
- Lookup micro-benchmarks (including longest-prefix-match validation)
- Memory footprint snapshots around bulk loads
- Parallel lookup benchmarks

### Notes on Scale Labels
- Benchmarks labeled “1M” operate on 1,000,000 prefixes.

### Reproducing
- Standard Go `testing` benchmarks are used. Run them with the usual `go test -bench=.` flags appropriate for your environment and version.

### Running benchmarks

- Run all benchmarks in the repo:

```bash
go test -bench=. -benchmem ./...
```

- Run only LPM benchmarks:

```bash
go test -bench='^BenchmarkLPM' -benchmem ./...
```

- Run only MapTrie benchmarks:

```bash
go test -bench='^BenchmarkMapTrie' -benchmem ./...
```

- Run only Patricia benchmarks:

```bash
go test -bench='^BenchmarkPatricia' -benchmem ./...
```

### Running the 1M benchmarks specifically

- Filter by function names that include "1M":

```bash
go test -bench='1M' -benchmem ./...
```

- Or target a specific suite’s 1M cases (examples):

```bash
# LPM 1M insert and lookup
go test -bench='^BenchmarkLPM(Insert1M|Lookup1M)$' -benchmem ./...

# MapTrie 1M insert and lookup
go test -bench='^BenchmarkMapTrie(Insert1M|Lookup1M)$' -benchmem ./...

# Patricia 1M insert and lookup
go test -bench='^BenchmarkPatricia(Insert1M|Lookup1M)$' -benchmem ./...
```