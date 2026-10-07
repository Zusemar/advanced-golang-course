# Week 1 demos

Six small programs, one per scheduler mechanism.

## Setup

- `go.mod` pins `go 1.27.1`. With `GOTOOLCHAIN=auto` (the default), any Go 1.21 or newer
  downloads 1.27.1 on first use, so you need network once. Older toolchains cannot run
  some demos: they use APIs from Go 1.25 and `runtime/metrics` names from Go 1.26.
- If your shell exports `GOROOT`, unset it. Otherwise the build fails with
  `compile: version ... does not match go tool version`.
- Build first, then run the binaries. `GODEBUG=... go run` also applies `GODEBUG` to the
  go command and the compiler, and their `SCHED` lines get mixed into yours.

```sh
mkdir -p /tmp/bin && go build -o /tmp/bin/ ./...
```

## Run

| Demo | Command | What it shows |
|---|---|---|
| 01-schedtrace | `GOMAXPROCS=2 GODEBUG=schedtrace=1000 /tmp/bin/01-schedtrace` | The fields of a schedtrace line. Add `,scheddetail=1` for a per-P, per-M, per-G dump. |
| 02-pingpong | `go test -run=X -bench=PingPong -benchtime=200000x ./02-pingpong` | runnext. With a CPU hog on the same P, ping-pong costs about 1.3 to 2.5x the plain run (281 vs 445 ns/op on an i5-13600KF). |
| 03-stealing | `GOMAXPROCS=1 GODEBUG=schedtrace=100 /tmp/bin/03-stealing` | The 256-slot local run queue overflowing to the global queue, and the 128-G refill batch. |
| 04-syscall | `GOMAXPROCS=2 /tmp/bin/04-syscall` | 200 goroutines blocked on the network cost at most one extra thread. 200 in a blocking syscall cost about 200 threads. |
| 05-preemption | `GOMAXPROCS=1 /tmp/bin/05-preemption`, then the same with `GODEBUG=asyncpreemptoff=1` | Async preemption frees main in 10 to 20 ms. With it off, the program hangs; press Ctrl-C. |
| 06-gomaxprocs | `/tmp/bin/06-gomaxprocs` and `taskset -c 0 /tmp/bin/06-gomaxprocs` | Where GOMAXPROCS comes from (the CPU affinity mask), and how to re-arm the automatic updates. |

Numbers depend on the machine. Yours will differ.

04-syscall does not build on macOS, which has no `syscall.Nanosleep`. Use Linux or a
container. The other five build on macOS; 06 needs `taskset`, which is Linux only.

## License

Apache-2.0, see [../../LICENSE-CODE](../../LICENSE-CODE). 02-pingpong adapts code from
`src/runtime/proc_test.go` in the Go distribution, BSD-3-Clause, (c) The Go Authors.
