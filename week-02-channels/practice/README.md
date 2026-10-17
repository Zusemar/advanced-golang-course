# Week 2 practice: fan-in, nil channels and drain-on-close

Merge several input channels into one: first with a goroutine per input and no `select`,
then with one `select` loop that is wrong on purpose, then fixed by switching off closed
inputs with `nil`. Then stop reading a merged channel early, watch goroutines leak, and fix
it by draining the channel to its close. Six small steps, each one already written. Every
run prints its own evidence: items checked, wall time against CPU time, goroutines left
behind.

| Step | Idea |
|---|---|
| step0-range | One input, read with `range` until it is closed. Every later step receives this way. |
| step1-goroutines | Fan-in without `select`: one goroutine per input copies it onto a shared output, and one more closes the output once they are all done. |
| step2-select-spin | Fan-in of two inputs with one `select` loop, wrong on purpose. A closed input is always ready to receive, so the loop spins until the other input closes too. |
| step3-select-nil | Step 2 with each closed input set to `nil`. A `nil` channel is never ready, so its case is switched off. |
| step4-early-exit | Take the first 10 items of a merge and stop reading. The right answer, and goroutines blocked forever. |
| step5-drain | Step 4 plus one loop that receives until the merge is closed. Nothing leaks; the price is waiting for the sources to finish. |

Step 1 and step 3 are then timed against each other.

Nothing here is one-to-many: no subscribe, no unsubscribe, no slow-subscriber policy. Those
are the homework, [`Broadcaster`](../homework/).

## Setup

- `go.mod` pins `go 1.27.1`. With `GOTOOLCHAIN=auto` (the default), any Go 1.21 or newer
  downloads 1.27.1 on first use, so you need network once. If you ran week 1 on this
  machine, you already have it.
- If your shell exports `GOROOT`, unset it. Otherwise the build fails with
  `compile: version ... does not match go tool version`.
- Run every command from inside this directory. Outside it, `go` may be an older
  toolchain.

## Run

```sh
sh run.sh
```

It checks that `go` here is go1.27.1, builds `bin/practice` and `bin/practice.test`, runs
`go vet` and `go test -race`, and prints every command of the session, in order. Run them
from this directory. The ones to try yourself:

```sh
bin/practice -step=2                       # the spin, on your machine
bin/practice -step=3
bin/practice -step=4
bin/practice -step=5
go test -race .                            # every step's check, and the leak check for steps 4 and 5
bin/practice -step=bench -cpu=1,2,8        # step 1 against step 3, timed: touch nothing while it runs
go test -run=X -bench=Merge -cpu=1,2,8 .   # the same comparison through go test
```

`bin/practice -h` lists the flags (`-items` and `-reps`, for `-step=bench` only).

If your machine has fewer than 8 CPUs, replace 8 with the NumCPU the header prints.

**Step 1, by hand.** Delete the line `close(out)` in `step1-goroutines/run.go` (it is
marked), rebuild with `go build -o bin/ ./cmd/practice`, and run `bin/practice -step=1`.
Then put the line back and rebuild (`git checkout step1-goroutines/run.go`, then the same
`go build`). Without that, step 5 and the bench deadlock too, and `sh run.sh` waits silently
for `go test`'s 10-minute timeout.

**Windows without a POSIX shell:** skip `run.sh`, run `go build -o bin/ ./cmd/practice`, then
`bin\practice.exe -step=2`. The CPU column reads `n/a` there, but the count of receives
from a closed input still shows the spin.

## What to expect

- Step 0: `len=5 cap=8` before the first receive. `range` still receives all 5 items after
  the close, then stops, and one more receive gives the zero value and `ok=false`.
- Step 1: 1070 items, `check ok`, CPU time a small fraction of wall time, and no goroutines
  left behind. Without the `close(out)` line:
  `fatal error: all goroutines are asleep - deadlock!`, with `goroutine 1 [chan receive]`
  in `Collect`, and exit status 2 (`echo $?`; the binary does not print it).
- Step 2: 1050 items and `check ok`, but CPU time about equal to wall time, and receives
  from a closed input in the millions. Only the CPU column shows the bug.
- Step 3: the same items and wall time as step 2, CPU time a small fraction of wall again,
  and exactly 2 receives from a closed input, one per input.
- Step 4: 10 items, fast, `check ok`, and 5 goroutines left behind. Nothing reports the
  leak.
- Step 5: the same 10 items, but `FirstK` returns only when both sources have finished, and
  no goroutines are left behind.
- The bench: one row per merge per `-cpu` value, with a median and `ns/item`, every row
  `check ok`.

Measured on 2026-10-16 on an i5-13600KF under WSL2, 20 logical CPUs, go1.27.1, five runs
of each step:

| Step | Wall | CPU | Left behind |
|---|---|---|---|
| 1-goroutines | 108 to 112 ms | 3 to 5 ms | 0 |
| 2-select-spin | 111 to 126 ms | 114 to 130 ms | (3.2 to 3.4 million receives from a closed input) |
| 3-select-nil | 108 ms | 2 to 3 ms | (2 receives from a closed input) |
| 4-early-exit | 5 to 6 ms | | 5 |
| 5-drain | 216 to 227 ms | | 0 |

The merge timing, `bin/practice -step=bench -cpu=1,2,8`, five runs, ns/item:

| cpu | 1-goroutines | 3-select-nil |
|---|---|---|
| 1 | 349 to 356 | 378 to 393 |
| 2 | 326 to 343 | 346 to 375 |
| 8 | 303 to 319 | 335 to 383 |

On that machine the goroutine per input was faster at every `-cpu` value, by about 6 to 14%.

Numbers depend on the machine. Yours will differ; the shape should not.

## License

Apache-2.0, see [../../LICENSE-CODE](../../LICENSE-CODE).
