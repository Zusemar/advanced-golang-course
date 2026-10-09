#!/bin/sh
# run.sh -- the week 1 practice session: build everything, then print
# the exact commands of the session, in order.
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Usage:  sh run.sh             build, vet, test, then print the session script
#         sh run.sh rehearse    the same, then run every timed command once and
#                               write the fallback traces fallback-step{3,4,5}.out
#                               (23 s on an i5-13600KF under WSL2,
#                               build cache warm; a noisy step 3 or 5 is re-recorded)
#         sh run.sh step        the same build, then walk the session script one
#                               command per key, for a presenter driving alone:
#                               Space runs it, b goes back one, s skips, q quits
#                               (Ctrl-C at the prompt does nothing). The "open" and
#                               "diff -u" hints are steps too, opened in VS Code
#                               when `code` is on PATH. Ctrl-C stops a
#                               running step, the trace UI, and returns here.
#                               After the last step only b (back to it) and q
#                               (quit) act; any other key, Ctrl-C included,
#                               shows the end prompt again. On a trace UI step, f
#                               opens the rehearsed fallback trace instead of the
#                               live one. The fallbacks are also served for the
#                               whole run, step N on 127.0.0.1:606N, so a second
#                               tab can compare two steps; q stops them.
#
# Works on Linux and macOS. Run it from anywhere; it cd's to its own directory.
# Nothing here is timed while it builds: build first, then measure.

set -eu
cd "$(dirname "$0")"

# The pin is enforced by the go directive in go.mod and go.work, not by the
# command name. Inside this repository an older `go` re-execs go1.27.1
# (GOTOOLCHAIN=auto, needs network the first time). Refuse to go on if it
# did not, because a go1.25 `go tool trace` cannot read a go1.27 trace.
v=$(go env GOVERSION)
if [ "$v" != "go1.27.1" ]; then
	echo "run.sh: go here is $v, want go1.27.1." >&2
	echo "run.sh: check GOTOOLCHAIN (want auto) and that this is inside the course repo." >&2
	exit 1
fi

if [ "${1:-}" = step ] && [ ! -t 0 ]; then
	echo "run.sh: step needs a terminal for the key prompt." >&2
	exit 1
fi

# An exported GOROOT survives the re-exec above: GOVERSION still says
# go1.27.1, but the compiler and go tool trace come from the older tree.
# Checked in every mode, before the build.
G=$(go env GOROOT)
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ] && [ -n "${GOROOT:-}" ]; then
	echo "run.sh: your shell exports GOROOT=$GOROOT, not the go1.27.1 tree; unsetting it for this run."
	unset GOROOT
	G=$(go env GOROOT)
fi
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "run.sh: go env GOROOT is $G, whose VERSION is not go1.27.1." >&2
	exit 1
fi

echo "== building with $v, $(go env GOOS)/$(go env GOARCH)"
go vet ./...
go build -o bin/ ./cmd/practice
go test -c -o bin/practice.test .
go test -race -count=1 . >/dev/null
echo "== built bin/practice and bin/practice.test; go vet and go test -race pass"

if [ "${1:-}" = "rehearse" ]; then
	echo "== rehearsal: the full table, then three traces. Touch nothing while it runs."
	bin/practice -cpu=1,2,8
	# The fallbacks have their own names, so a live trace in the session cannot
	# overwrite them. A busy box spoils a trace: workers start late, "peak
	# running" stays under 8 and the staircase blurs. So each fallback is the
	# fastest of five recordings that had all 8 workers running at once.
	for s in 3 4 5; do
		best=
		for try in 1 2 3 4 5; do
			line=$(bin/practice -step=$s -cpu=8 -trace=fallback-step$s.try.out | sed -n 2p)
			us=$(printf '%s\n' "$line" | awk 'noisy($0) { next } {
				t = $4; u = 1000
				if (t ~ /ms$/) sub(/ms$/, "", t); else if (t ~ /s$/) { sub(/s$/, "", t); u = 1000000 }
				printf "%d", t * u }
				function noisy(l,  c, p) { c = l; sub(/.*cpu=/, "", c); sub(/ .*/, "", c)
					p = l; sub(/.*peak running /, "", p); sub(/,.*/, "", p); return p + 0 < c + 0 }')
			if [ -n "$us" ] && { [ -z "$best" ] || [ "$us" -lt "$best" ]; }; then
				best=$us
				mv fallback-step$s.try.out fallback-step$s.out
			fi
		done
		rm -f fallback-step$s.try.out
		if [ -z "$best" ]; then
			echo "== step $s: no clean recording in five (peak running under 8). The box is busy; fallback-step$s.out not replaced." >&2
			continue
		fi
		echo "== fallback-step$s.out: $((best / 1000)) ms, fastest clean run of five"
		go tool trace -pprof=sched fallback-step$s.out >sched-step$s.prof
	done
	for s in 4 5; do
		echo "== scheduler latency, step $s (time runnable but not running):"
		go tool pprof -top -nodecount=4 sched-step$s.prof 2>/dev/null | sed -n '4,8p'
	done
	echo "== rehearsal done. fallback-step{3,4,5}.out are the fallback traces; step mode opens them with f."
	exit 0
fi

# The session script: printed as it stands by default, and walked line by line
# by step mode, so the two cannot drift apart.
script_text() {
	cat <<'EOF'

Week 1 practice session script. Every command runs from this directory.

  # 0. Tier 0, the baseline and the oracle           open step0-sequential/run.go
  bin/practice -step=0 -cpu=1

  # 1. one goroutine per task                        open step1-unbounded/run.go
  bin/practice -step=1 -cpu=1,2,8

  # 2. at most n at once                             diff -u step1-unbounded/run.go step2-semaphore/run.go
  bin/practice -step=2 -cpu=1,2,8
  bin/practice -step=2 -cpu=8 -n=2                   # the bound is real

  # 3. n workers, one chunk each                     open step3-chunked/run.go
  bin/practice -step=3 -cpu=1,2,8
  bin/practice -step=3 -cpu=8 -trace=trace-step3.out
  go tool trace -http=127.0.0.1:6060 trace-step3.out

  # 4. n workers, one shared queue                   diff -u step3-chunked/run.go step4-workerpool/run.go
  bin/practice -step=4 -cpu=1,2,8
  bin/practice -step=4 -cpu=8 -trace=trace-step4.out
  go tool trace -http=127.0.0.1:6060 trace-step4.out

  # 5. fill the queue first                          diff -u step4-workerpool/run.go step5-prefilled/run.go
  bin/practice -step=5 -cpu=1,2,8

  # the table (21 s on an i5-13600KF; touch nothing)
  bin/practice -cpu=1,2,8

  # the same through go test, which is what students run
  bin/practice.test -test.run=X -test.bench=Run -test.benchtime=3x -test.cpu=1,2,8

Fewer than 8 CPUs on this machine? Replace 8 with the NumCPU the header prints.
A noisy live trace (workers with milliseconds of Sched wait): open the rehearsed
fallback-step3.out or fallback-step4.out instead (f in step mode).
A trace UI that will not open: go tool trace -pprof=sched fallback-step4.out > s.prof
&& go tool pprof -top s.prof  (the same finding, as text). Ctrl-C stops the UI.
EOF
}

if [ "${1:-}" = step ]; then
	# One record per step, kind|section header|command. Indented "# ..." lines
	# are section headers; an "open FILE" or "diff -u A B" at the end of one is
	# a step of its own. Other indented lines are commands, run as printed.
	steps=$(script_text | awk '
		/^  #/ {
			ctx = substr($0, 3)
			if ((i = index(ctx, "   open ")) > 0) k = "open"
			else if ((i = index(ctx, "   diff -u ")) > 0) k = "diff"
			else next
			h = substr(ctx, i); sub(/^ +/, "", h)
			# The command prints on its own line below; keep the header short.
			ctx = substr(ctx, 1, i - 1); sub(/ +$/, "", ctx)
			print k "|" ctx "|" h
			next
		}
		/^  [^ ]/ { print "run|" ctx "|" substr($0, 3) }')
	nsteps=$(printf '%s\n' "$steps" | wc -l | tr -d ' ')
	saved=$(stty -g)
	# The rehearsed fallbacks stay up for the whole session, step N on port
	# 606N, so a live trace on 6060 can be compared with another step's in a
	# second tab: Ctrl-C stops a trace server and its tabs with it. They are
	# background jobs of a script, so Ctrl-C does not reach them; they stop when
	# this script exits. BROWSER=true keeps each from opening a tab of its own.
	# The tool binary runs directly, so $! is the server and not a go wrapper.
	tracetool=$(go tool -n trace)
	pids=
	urls=
	for s in 3 4 5; do
		[ -f fallback-step$s.out ] || continue
		BROWSER=true "$tracetool" -http=127.0.0.1:606$s fallback-step$s.out >/dev/null 2>&1 &
		pids="$pids $!"
		urls="$urls  step $s http://127.0.0.1:606$s"
	done
	trap 'stty "$saved" 2>/dev/null; [ -z "$pids" ] || kill $pids 2>/dev/null' EXIT
	trap 'exit 129' HUP
	trap 'exit 143' TERM
	# A child of a shell that ignores INT inherits SIG_IGN, so this is a
	# handler, not trap '' INT: Ctrl-C kills the running step, not this script.
	trap : INT
	# key: one keypress, no echo, no Enter. First drain keys pressed while a
	# step ran (min 0 time 0 returns at once), so they cannot start the next.
	key() {
		stty raw -echo min 0 time 0
		dd bs=256 count=1 >/dev/null 2>&1 || :
		stty raw -echo min 1 time 0
		k=$(dd bs=1 count=1 2>/dev/null)
		stty "$saved"
	}
	script_text | sed -n '2,3p'
	[ -z "$urls" ] || echo "== fallback traces, up until q, for comparison:$urls"
	echo "== step mode, $nsteps steps. Keys: Space run, b back, s skip, q quit, f fallback trace (Ctrl-C at the prompt does nothing)"
	i=1
	while :; do
		if [ "$i" -gt "$nsteps" ]; then
			echo
			script_text | sed -n '/^Fewer than/,$p'
			# Only b and q leave the end prompt: a stray key, Ctrl-C included,
			# shows it again rather than ending the script after the last step.
			while :; do
				echo "== end of the script. b goes back to the last step; q quits."
				key
				case $k in
				b)
					i=$nsteps
					break
					;;
				q) exit 0 ;;
				esac
			done
			continue
		fi
		IFS='|' read -r kind ctx cmd <<-EOS
		$(printf '%s\n' "$steps" | sed -n "${i}p")
		EOS
		echo
		echo "-- step $i of $nsteps   $ctx"
		case $kind in
		open) echo "   $cmd" ;;
		*) echo "   \$ $cmd" ;;
		esac
		# On a trace UI step, f opens the rehearsed fallback instead.
		fb=
		case $cmd in
		"go tool trace -http="*)
			echo "   Ctrl-C (left hand) to stop the UI"
			[ -z "$urls" ] || echo "   compare with the fallbacks:$urls"
			fb=fallback-${cmd##*trace-}
			;;
		esac
		case $cmd in
		"bin/practice -step=5 -cpu=1,2,8")
			[ -z "$urls" ] || echo "   then compare main.main, step 4 against step 5: http://127.0.0.1:6064 and http://127.0.0.1:6065, Goroutine analysis"
			;;
		esac
		[ "$kind" = diff ] && ! command -v code >/dev/null 2>&1 && echo "   (paged if long: q comes back)"
		if [ -n "$fb" ] && [ -f "$fb" ]; then
			echo "   [Space] run  [f] open $fb instead  [b] back  [s] skip  [q] quit"
		elif [ -n "$fb" ]; then
			echo "   [Space] run  [b] back  [s] skip  [q] quit   (no $fb: sh run.sh rehearse)"
		else
			echo "   [Space] run  [b] back  [s] skip  [q] quit"
		fi
		key
		case $k in
		' ') ;;
		f)
			[ -n "$fb" ] && [ -f "$fb" ] || continue
			cmd="${cmd%trace-*}$fb"
			echo "   \$ $cmd"
			;;
		b)
			[ "$i" -gt 1 ] && i=$((i - 1))
			continue
			;;
		s)
			i=$((i + 1))
			continue
			;;
		q)
			echo "== quit at step $i."
			exit 0
			;;
		*) continue ;; # Ctrl-C (\003) and any other key: show the step again
		esac
		rc=0
		case $kind in
		open)
			f=${cmd#open }
			if command -v code >/dev/null 2>&1 && code -g "$PWD/$f"; then
				echo "   opened in VS Code: $f"
			elif command -v vim >/dev/null 2>&1; then
				echo "   vim, read-only: ZZ (left Shift+Z twice) to come back"
				vim -M "$f" || echo "   no editor here: open it in the host editor"
			else
				echo "   no editor here: open it in the host editor"
			fi
			;;
		diff)
			set -- ${cmd#diff -u }
			if command -v code >/dev/null 2>&1 && code --diff "$PWD/$1" "$PWD/$2"; then
				echo "   opened in VS Code: $1 vs $2"
			elif command -v less >/dev/null 2>&1; then
				eval "$cmd" | less -FX || rc=$?
			else
				eval "$cmd" || :
			fi
			;;
		*)
			case $cmd in
			*" -trace="*)
				# A live trace recorded on a busy box: say so before the UI step.
				out=$(eval "$cmd") || rc=$?
				printf '%s\n' "$out"
				printf '%s\n' "$out" | awk '/peak running/ {
					c = $0; sub(/.*cpu=/, "", c); sub(/ .*/, "", c)
					p = $0; sub(/.*peak running /, "", p); sub(/,.*/, "", p)
					if (p + 0 < c + 0) print "== noisy: peak running " p " of " c ", so this trace will not be clean. b records again; f at the next step opens the fallback." }'
				;;
			*) eval "$cmd" || rc=$? ;;
			esac
			;;
		esac
		case $rc:$cmd in
		0:*) ;;
		*:"go tool trace -http="*) echo "== trace UI stopped (go tool trace exits $rc on Ctrl-C)" ;;
		*) echo "== exit status $rc" ;;
		esac
		i=$((i + 1))
	done
fi

script_text
