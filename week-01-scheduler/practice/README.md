# Week 1 practice: a bounded parallel executor

`Run(n, tasks)`: run every task, with no more than `n` running at once. Built in six
small steps, each one already written, each measured at `-cpu=1,2,8` against step 0.

| Step | Idea |
|---|---|
| step0-sequential | One loop. The baseline, and the oracle every other step is checked against. |
| step1-unbounded | One goroutine per task. Fast, but `n` is ignored: 1000 tasks are 1000 goroutines. |
| step2-semaphore | Still one goroutine per task, but a buffered channel as a semaphore lets at most `n` run a task at once. |
| step3-chunked | `n` workers, one contiguous chunk each. With uneven tasks, workers that finish early sit idle. |
| step4-workerpool | `n` workers pulling from one shared queue. |
| step5-prefilled | The same, with the queue filled before the workers start. |

Ordered results, error propagation and cancellation are not built here. Those are the
homework.

## Setup

- `go.mod` pins `go 1.27.1`. With `GOTOOLCHAIN=auto` (the default), any Go 1.21 or newer
  downloads 1.27.1 on first use, so you need network once.
- If your shell exports `GOROOT`, unset it. Otherwise the build fails with
  `compile: version ... does not match go tool version`.
- Run `go tool trace` from inside this directory. Outside it, `go` may be an older
  toolchain, and it cannot read a go1.27 trace.

## Run

```sh
sh run.sh
```

It builds `bin/practice` and `bin/practice.test`, runs `go vet` and `go test -race`, and
prints every command of the session, in order. Run them from this directory. The most
useful ones:

```sh
bin/practice -step=1 -cpu=1,2,8                      # one step, one row per GOMAXPROCS
bin/practice -step=2 -cpu=8 -n=2                     # the bound is real: about 2x
bin/practice -step=3 -cpu=8 -trace=trace-step3.out   # then:
go tool trace -http=127.0.0.1:6060 trace-step3.out
bin/practice -cpu=1,2,8                              # all six steps
```

`sh run.sh step` walks the same commands one key at a time: Space runs, b goes back,
s skips, q quits. `bin/practice -h` lists the flags (`-tasks`, `-shape=even`, `-reps`).

If your machine has fewer than 8 CPUs, replace 8 with the NumCPU the header prints.

## What to expect

Speedup over step 0, measured on an i5-13600KF under WSL2, go1.27.1, ten runs:

| Step | cpu=2 | cpu=8 |
|---|---|---|
| 1-unbounded | 1.93 to 2.03x | 7.40 to 7.54x |
| 2-semaphore | 1.93 to 2.03x | 7.38 to 7.52x |
| 3-chunked | 1.29 to 1.35x | 3.88 to 4.28x |
| 4-workerpool | 1.78 to 1.88x | 6.39 to 6.51x |
| 5-prefilled | 1.96 to 2.04x | 7.44 to 7.56x |

Numbers depend on the machine. Yours will differ; the shape should not. At `-cpu=1`
every step is within noise of 1.00x.

## License

Apache-2.0, see [../../LICENSE-CODE](../../LICENSE-CODE).
