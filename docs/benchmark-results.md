# Benchmark Results

## Окружение
- OS: darwin
- Arch: arm64
- CPU: Apple M4
- Package: `github.com/HAL-X9/search-trends-service/internal/usecases`

## Команды

```bash
go test ./internal/usecases -run ^$ -bench . -benchmem -count=3
```

## Результаты

```
BenchmarkTrendsInteractor_ProcessQuery-10     11656549    101.6 ns/op    32 B/op    1 allocs/op
BenchmarkTrendsInteractor_ProcessQuery-10     11860592    101.6 ns/op    32 B/op    1 allocs/op
BenchmarkTrendsInteractor_ProcessQuery-10     11802406    101.5 ns/op    32 B/op    1 allocs/op

BenchmarkTrendsInteractor_GetTopTrends-10     100000000    14.68 ns/op     8 B/op    0 allocs/op
BenchmarkTrendsInteractor_GetTopTrends-10     100000000    14.76 ns/op     8 B/op    0 allocs/op
BenchmarkTrendsInteractor_GetTopTrends-10     100000000    14.71 ns/op     8 B/op    0 allocs/op

BenchmarkSlidingWindow_AggregateAll-10          482826    2472 ns/op      256 B/op    2 allocs/op
BenchmarkSlidingWindow_AggregateAll-10          486255    2544 ns/op      256 B/op    2 allocs/op
BenchmarkSlidingWindow_AggregateAll-10          486091    2458 ns/op      256 B/op    2 allocs/op
```

## Резюме

```
ProcessQuery: ~101.5 ns/op, 1 alloc/op.
GetTopTrends: ~14.7 ns/op, 0 alloc/op.
AggregateAll: ~2.5 µs/op, 2 alloc/op.
```