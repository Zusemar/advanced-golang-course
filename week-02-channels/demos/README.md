# Week 2 demos

Eight small programs and benchmarks, one per claim about channels and `select`.

## Setup

- `go.mod` pins `go 1.27.1`. With `GOTOOLCHAIN=auto` (the default), any Go 1.21 or newer
  downloads 1.27.1 on first use, so you need network once.
- If your shell exports `GOROOT`, unset it. Otherwise the build fails with
  `compile: version ... does not match go tool version`.
- Run every command from this directory, the module root.

## Run

| Demo | Command | What it shows |
|---|---|---|
| 04-nilcase | `go test -run=X -bench=. -benchtime=2000000x -cpu=1 ./04-nilcase` | One sixteen-case `select`, one case ready. Setting the other fifteen channels to `nil` makes it several times faster: `selectgo` leaves nil channels out of the poll order and the lock order. |
| 01-handoff | `go test -run=X -bench=. -benchtime=300000x -count=5 -cpu=1,2,8 ./01-handoff` | What a buffer buys. Request/response (`pingpong`) gains nothing from it; one-way flow at different speeds (`stream`) does. Run with `-count=5`: single runs of the ping-pong pair overlap. |
| 02-select | `go test -run=X -bench=. -benchtime=1000000x -count=3 -cpu=1 ./02-select` | What a `select` costs by number of cases. A one-case `select` is a plain receive; `select` with `default` never takes the channel lock. Each iteration also does one buffered send; subtract it before comparing rows. |
| 08-lockcost | `go test -run=X -bench=. -benchtime=1000000x -count=3 -cpu=1 ./08-lockcost` | Where the `select` time goes. Same sixteen cases, but all on one channel, so `sellock` takes one lock instead of sixteen. The difference is the locking. |
| 03-fairness | `go run ./03-fairness` | `select` picks uniformly among the ready cases: three numbers near 33.3%, not exactly equal. |
| 05-semantics | `go run ./05-semantics` | Every rule for closed and nil channels, executed rather than recited, plus a timer channel holding a value while reporting `len=0 cap=0`. |
| 06-lencap | `go run ./06-lencap`, then `go run -race ./06-lencap` | A `len()` poll that looks like synchronisation and is a data race, and a `close` versus `len` race the detector ignores on purpose. Under `-race` the program exits 66 (`go run` prints `exit status 66`): that is the demo. |
| 07-waitreasons | `go run ./07-waitreasons` | Six goroutines blocked six different ways, then a deadlock dump. Read the wait reason in brackets on each goroutine. The program exits 2 by design (`go run` prints `exit status 2`). |

Numbers depend on the machine. Yours will differ; compare ratios, not nanoseconds.

All eight build on Linux and macOS.

## License

Apache-2.0, see [../../LICENSE-CODE](../../LICENSE-CODE).
