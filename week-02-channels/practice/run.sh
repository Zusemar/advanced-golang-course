#!/bin/sh
# run.sh -- the week 2 practice session: build everything, then print
# the exact commands of the session, in order.
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Usage:  sh run.sh             build, vet, test, then print the session script
#         sh run.sh rehearse    the same, then run every step once and the
#                               merge timing
#
# Works on Linux and macOS. Run it from anywhere; it cd's to its own directory.
# Nothing here is timed while it builds: build first, then measure.

set -eu
cd "$(dirname "$0")"

# The pin is enforced by the go directive in go.mod and go.work, not by the
# command name. Inside this repository an older `go` re-execs go1.27.1
# (GOTOOLCHAIN=auto, needs network the first time). Refuse to go on if it
# did not.
v=$(go env GOVERSION)
if [ "$v" != "go1.27.1" ]; then
	echo "run.sh: go here is $v, want go1.27.1." >&2
	echo "run.sh: check GOTOOLCHAIN (want auto) and that this is inside the course repo." >&2
	exit 1
fi

echo "== building with $v, $(go env GOOS)/$(go env GOARCH)"
go vet ./...
go build -o bin/ ./cmd/practice
go test -c -o bin/practice.test .
go test -race -count=1 . >/dev/null
echo "== built bin/practice and bin/practice.test; go vet and go test -race pass"

if [ "${1:-}" = "rehearse" ]; then
	echo "== rehearsal: every step, then the merge timing. Touch nothing while it runs."
	for s in 0 1 2 3 4 5; do
		echo "== step $s"
		bin/practice -step=$s
	done
	bin/practice -step=bench -cpu=1,2,8
	echo "== rehearsal done."
	exit 0
fi

cat <<'EOT'

Week 2 practice session script. Every command runs from this directory.

  # 0. range drains a closed channel                 open step0-range/run.go
  bin/practice -step=0

  # 1. fan-in without select                         open step1-goroutines/run.go
  bin/practice -step=1

  # 2. fan-in with select, wrong on purpose          open step2-select-spin/run.go
  bin/practice -step=2

  # 3. switch off a closed input with nil            diff -u step2-select-spin/run.go step3-select-nil/run.go
  bin/practice -step=3

  # 4. stop reading early                            open step4-early-exit/run.go
  bin/practice -step=4

  # 5. drain to close                                diff -u step4-early-exit/run.go step5-drain/run.go
  bin/practice -step=5

  # step 1 against step 3, timed (touch nothing)
  bin/practice -step=bench -cpu=1,2,8

  # the same through go test, which is what students run
  bin/practice.test -test.run=X -test.bench=Merge -test.cpu=1,2,8

Fewer than 8 CPUs on this machine? Replace 8 with the NumCPU the header prints.
EOT
