#!/bin/sh
# cues.sh -- L-02, the week 2 lecture: the run cues of demos.md "Run order for
# the lecture" and the go-to-line stops of the deck, as a runner for a presenter
# who drives alone with one hand. One key per cue: nothing is typed during the
# lecture. Copied from week 1's cues.sh: the CUES table, the cache warm-up and
# --present are this week's; the pin and GOROOT checks are week 1's, unchanged.
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Usage:  sh cues.sh            warm the build cache, then walk the cues one key each
#         sh cues.sh --solo     the same, without the cues on cut-tagged slides
#         sh cues.sh --short    the 60-minute cut (slides.ru.60.html): run cues
#                               D4, D8, D6r and twelve open cues
#         sh cues.sh --present  hide the Expect lines and clear the screen at each
#                               cue, for a shared screen; combines with the others
#         sh cues.sh --list     print the cues and run nothing (the transcript
#                               fallback); combines with --solo and --short
#         sh cues.sh --help     this text
#
# Keys at each cue: Space runs it, b goes back one cue, s skips it, q quits.
# Ctrl-C at the key prompt does nothing; only q quits. Ctrl-C while a cue
# runs stops that command and returns here. Keys pressed while a cue runs
# are discarded. After the last cue only b (back to it) and q (quit) act.
#
# Run cues are the commands of demos.md, verbatim. D6r and D7 exit non-zero
# by design (go run reports "exit status 66" and "exit status 2"): say so
# before running them. Open cues open the pinned source at FILE:LINE.
# Run it from anywhere; it cd's to its own directory, the demos module.

set -eu
cd "$(dirname "$0")"

usage() {
	sed -n '11,26s/^# \{0,1\}//p' "$(basename "$0")"
}

solo=0
short=0
list=0
present=0
for a in "$@"; do
	case $a in
	--solo) solo=1 ;;
	--short) solo=1 short=1 ;;
	--list) list=1 ;;
	--present) present=1 ;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "cues.sh: unknown argument: $a" >&2
		usage >&2
		exit 2
		;;
	esac
done

# The cues, derived from the deck (slides.ru.html, 64 slides, 2026-10-12), in
# deck order: slide 3 D4, 7 O1, 13 O3, 15 O4, 18 D1, 21 O2, 23 O5, 25 O6, 26 O7,
# 30 O8, 32 D2, 34 O10, 37 D3, 38 O11, 39 O13, 43 D8, 49 D5, 52 O12, 55 D7,
# 56 O14, 59 D6r. "cut" marks a cue whose slide carries a cut tag. Fields: id | lecture section | cut (cuttable) or - |
# run or open | the command, or FILE:LINE to open | what to expect.
# $G stands for $(go env GOROOT), resolved in this directory. Open-cue lines
# were checked against the go1.27.1 tree on 2026-10-12; Expect quotes the line.
# Run-cue Expect lines describe shape, not figures: re-measure before quoting one.
CUES='D4|2|-|run|go test -run=X -bench=. -benchtime=2000000x -cpu=1 ./04-nilcase|two lines: BenchmarkAllLive several times BenchmarkFifteenNil (WSL box 2026-10-12, n=10 median ≈ 11.9x (~12x))
O1|3.1|-|open|$G/src/runtime/chan.go:34|type hchan struct {
O3|5.1|-|open|$G/src/runtime/chan.go:176|func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
O4|5.2|-|open|$G/src/runtime/chan.go:318|func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
D1|5.6|cut|run|go test -run=X -bench=. -benchtime=300000x -cpu=1,2,8 ./01-handoff|18 lines; Stream64 and Stream1024 several times faster than Stream0; Unbuffered and Buffered1 close to each other
O2|3.3|-|open|$G/src/runtime/runtime2.go:404|type sudog struct {
O5|5.4|-|open|$G/src/runtime/chan.go:392|func sendDirect(t *_type, sg *sudog, src unsafe.Pointer) {
O6|5.5|-|open|$G/src/runtime/chan.go:524|func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
O7|5.5|cut|open|$G/src/runtime/chan.go:702|func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
O8|6.1|-|open|$G/src/cmd/compile/internal/walk/select.go:37|// optimization: zero-case select (lines 37-72: the 0- and 1-case rewrites)
D2|6.1|cut|run|go test -run=X -bench=. -benchtime=1000000x -count=3 -cpu=1 ./02-select|three lines per benchmark; Select1 about PlainRecv; Select2..16 grow faster than linear; SelectDefault and SelectClosed a few ns
O10|6.2|-|open|$G/src/runtime/select.go:167|// generate permuted order
D3|6.2|cut|run|go run ./03-fairness|three cases at about 33.3% each of 300000
O11|6.3|-|open|$G/src/runtime/select.go:206|// sort the cases by Hchan address to get the locking order.
O13|6.4|-|open|$G/src/runtime/select.go:264|// pass 1 - look for something already waiting
D8|6.5|-|run|go test -run=X -bench=. -benchtime=1000000x -count=3 -cpu=1 ./08-lockcost|SixteenChans about 3.9x OneChanSixteenCases (WSL 2026-10-12, n=10: 492 vs 126), three lines each
D5|7.1|cut|run|go run ./05-semantics|eight rules, each "ok" or the panic it names (13 lines with the -> lines); the time.Timer rule shows len=0 cap=0 (buffer holds a value)
O12|7.2|-|open|$G/src/runtime/chan.go:414|func closechan(c *hchan) {
D7|7.5|cut|run|go run ./07-waitreasons|fatal error: all goroutines are asleep - deadlock! then seven goroutines: select (no cases), chan receive (nil chan), chan send (nil chan), chan receive, chan send, select (no cases), select; exit status 2, then == exit status 1 (from go run)
O14|7.5|-|open|$G/src/runtime/proc.go:6511|// There are no goroutines running, so we can look at the P'"'"'s.
D6r|8|-|run|go run -race ./06-lencap|one WARNING: DATA RACE at main.go:28 vs main.go:22; close vs len/cap: no report; Found 1 data race(s); exit status 66, then == exit status'



CUES_ALL=$CUES
if [ "$solo" = 1 ]; then
	CUES=$(printf '%s\n' "$CUES" | grep -v '^[^|]*|[^|]*|cut|')
fi
if [ "$short" = 1 ]; then
	# The 60-minute cut: the cues on slides the outline's 60 column keeps
	# (w2-outline.md, 2026-10-12), checked against slides.ru.60.html: the cues on
	# its 40 kept slides (the "full deck: slide N" comments). Taken from the full table.
	CUES=$(printf '%s\n' "$CUES_ALL" | grep -E '^(D4|O1|O3|O4|O2|O5|O6|O8|O10|O11|O13|D8|O12|O14|D6r)\|')
fi
ncues=$(printf '%s\n' "$CUES" | wc -l | tr -d ' ')

# show N: print cue N (a line number in $CUES) and set id, kind, arg.
show() {
	IFS='|' read -r id sect cut kind arg expect <<-EOF
	$(printf '%s\n' "$CUES" | sed -n "${1}p")
	EOF
	tag=
	[ "$cut" = cut ] && tag='   (cuttable)'
	if [ "$present" = 1 ] && [ "$list" = 0 ]; then
		printf '\033[2J\033[H'
	else
		echo
	fi
	echo "== cue $id   [§$sect]   $1 of $ncues$tag"
	if [ "$kind" = open ]; then
		echo "   open $arg"
	else
		echo "   \$ $arg"
	fi
	[ "$present" = 1 ] && [ "$list" = 0 ] && return 0
	echo "   Expect: $expect"
}

if [ "$list" = 1 ]; then
	echo "Cues for L-02, from demos.md and the deck. Run from this directory."
	echo "\$G is \$(go env GOROOT), run here. Warm the build cache first: go vet ./..."
	i=1
	while [ "$i" -le "$ncues" ]; do
		show "$i"
		i=$((i + 1))
	done
	exit 0
fi

if [ ! -t 0 ]; then
	echo "cues.sh: needs a terminal for the key prompt; use --list to print the cues." >&2
	exit 1
fi

# The pin is enforced by the go directive in go.mod and go.work, not by the
# command name. Inside this repository an older `go` re-execs go1.27.1
# (GOTOOLCHAIN=auto, needs network the first time). Refuse to go on if it
# did not, because the cues quote go1.27.1 source lines and output.
v=$(go env GOVERSION)
if [ "$v" != "go1.27.1" ]; then
	echo "cues.sh: go here is $v, want go1.27.1." >&2
	echo "cues.sh: check GOTOOLCHAIN (want auto) and that this is inside the course repo." >&2
	exit 1
fi

# An exported GOROOT that names an older tree. When an older `go` re-execs
# go1.27.1, the toolchain switch resets GOROOT to the pinned tree and all is
# well. When the go1.27.1 binary is first on PATH there is no switch and the
# exported GOROOT wins: GOVERSION still says go1.27.1, but go env GOROOT
# (the open cues' source), go list, go doc and gopls all resolve to the older tree,
# and a build fails with
#   compile: version "go1.25.3" does not match go tool version "go1.27.1"
G=$(go env GOROOT)
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ] && [ -n "${GOROOT:-}" ]; then
	echo "cues.sh: your shell exports GOROOT=$GOROOT, not the go1.27.1 tree; unsetting it for this run."
	unset GOROOT
	G=$(go env GOROOT)
fi
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "cues.sh: go env GOROOT is $G, whose VERSION is not go1.27.1." >&2
	exit 1
fi

# A GODEBUG or GOMAXPROCS left in the shell would leak into every cue.
for e in GODEBUG GOMAXPROCS; do
	if eval "[ -n \"\${$e+x}\" ]"; then
		echo "cues.sh: unsetting $e, which your shell exports; the cues set their own."
		unset "$e"
	fi
done

tmp=$(mktemp "${TMPDIR:-/tmp}/cues.XXXXXX")
saved=$(stty -g)
trap 'stty "$saved" 2>/dev/null; rm -f "$tmp"' EXIT
trap 'exit 129' HUP
trap 'exit 143' TERM

# Warm the build cache, so the first go run and go test are not slow on screen.
# vet compiles every package, test-only ones included.
echo "== warming the build cache with $v, $(go env GOOS)/$(go env GOARCH): go vet ./..."
if ! go vet ./... >"$tmp" 2>&1; then
	cat "$tmp" >&2
	echo "cues.sh: go vet failed." >&2
	exit 1
fi
echo "== ready"
echo "== keys: Space run, b back, s skip, q quit (Ctrl-C at the prompt does nothing)"

# key: wait for one keypress, without echo or Enter, and set k to it. Keys
# pressed while a cue ran are already buffered: drain them first (min 0
# time 0 makes the read return at once), so a stray Space cannot start the
# next cue.
key() {
	stty raw -echo min 0 time 0
	dd bs=256 count=1 >/dev/null 2>&1 || :
	stty raw -echo min 1 time 0
	k=$(dd bs=1 count=1 2>/dev/null)
	stty "$saved"
}

# A child of a shell that ignores INT inherits SIG_IGN, so this is a handler,
# not trap '' INT: Ctrl-C kills the running cue and the runner carries on.
trap : INT

# openat FILE LINE: in VS Code if its CLI works, otherwise read-only vim.
openat() {
	if command -v code >/dev/null 2>&1 && code -g "$1:$2"; then
		echo "   opened in VS Code: $1:$2"
	elif command -v vim >/dev/null 2>&1; then
		echo "   vim, read-only: ZZ (left Shift+Z twice) to come back"
		vim -M "+$2" "$1" || echo "   no editor here: open it in the host editor"
	else
		echo "   no editor here: open it in the host editor"
	fi
	[ "$present" = 1 ] || echo "   line $2 reads: $(sed -n "${2}p" "$1")"
}

i=1
while :; do
	if [ "$i" -gt "$ncues" ]; then
		echo
		# Only b and q leave the end prompt: a stray key, Ctrl-C included,
		# shows it again rather than ending the runner after the last cue.
		while :; do
			echo "== end of cues. b goes back to the last one; q quits."
			key
			case $k in
			b)
				i=$ncues
				break
				;;
			q) exit 0 ;;
			esac
		done
		continue
	fi
	show "$i"
	echo "   [Space] run  [b] back  [s] skip  [q] quit"
	key
	case $k in
	' ') ;;
	b)
		[ "$i" -gt 1 ] && i=$((i - 1))
		continue
		;;
	s)
		i=$((i + 1))
		continue
		;;
	q)
		echo "== quit at cue $id."
		exit 0
		;;
	*) continue ;; # Ctrl-C (\003) and any other key: show the cue again
	esac
	if [ "$kind" = open ]; then
		case $arg in
		'$G'/*) f="$G${arg#'$G'}" ;;
		*) f="$PWD/$arg" ;;
		esac
		openat "${f%:*}" "${f##*:}"
	else
		rc=0
		eval "$arg" || rc=$?
		case $rc in
		0) ;;
		130) echo "== stopped by Ctrl-C (exit status 130)" ;;
		*) echo "== exit status $rc" ;;
		esac
	fi
	i=$((i + 1))
done
