#!/bin/sh
# cues.sh -- L-03, the week 3 lecture: the run cues of demos.md "Run order for
# the lecture", as a runner for a presenter who drives alone with one hand. One
# key per cue: nothing is typed during the lecture. Copied from week 2's
# cues.sh: the CUES table, the cache warm-up, --short and the single-cue mode are
# this week's; the pin and GOROOT checks and the key loop are week 2's.
#
# SPDX-License-Identifier: Apache-2.0
# Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
#
# Usage:  sh cues.sh            warm the build cache, then walk the cues one key each
#         sh cues.sh --short    the 60-minute deck's live cues only: D5p1, D5, D1,
#                               D1p1, D1p1r, D4-04, D4-05, D8, D3, D7 (D1r, D6
#                               and D2 are full-deck only)
#         sh cues.sh --present  hide the Expect and fallback lines and clear the
#                               screen at each cue, for a shared screen
#         sh cues.sh --list     print the cues, their expected output and the
#                               fallback line, and run nothing; combines with --short
#         sh cues.sh ID         run one cue once and exit, e.g. sh cues.sh D4-05
#         sh cues.sh --help     this text
#
# Keys at each cue: Space runs it, b goes back one cue, s skips it, q quits.
# Ctrl-C at the key prompt does nothing; only q quits. Ctrl-C while a cue
# runs stops that command and returns here. Keys pressed while a cue runs
# are discarded. After the last cue only b (back to it) and q (quit) act.
#
# Run cues are the commands of demos.md, verbatim. D1p1r, D1r and D4-02, -05, -10
# exit non-zero by design: the program exits 66, go run prints "exit status 66"
# and itself exits 1. Say so before running them. Never add -race to D3.
# Run it from anywhere; it cd's to its own directory, the demos module.

set -eu
cd "$(dirname "$0")"

usage() {
	sed -n '11,25s/^# \{0,1\}//p' "$(basename "$0")"
}

short=0
list=0
present=0
only=
for a in "$@"; do
	case $a in
	--short) short=1 ;;
	--list) list=1 ;;
	--present) present=1 ;;
	-h | --help)
		usage
		exit 0
		;;
	-*)
		echo "cues.sh: unknown argument: $a" >&2
		usage >&2
		exit 2
		;;
	*) only=$a ;;
	esac
done

# Where the Expect lines come from. Every range was measured on this machine:
# re-measure before quoting it on another.
STAMP='WSL2 on i5-13600KF, 20 logical CPUs, go1.27.1, 2026-10-18 (host CPU 9-18% busy); D5p1 and D5 on 2026-10-21 (host not checked for quiet; a browser and OBS running)'

# The cues, in lecture order (demos.md "Run order for the lecture"). Fields:
# id | lecture section | short (in the 60-minute deck) or - | the command |
# what to expect, as measured | the line to say if the run falls outside it.
CUES='D5p1|2, 9.3|short|GOMAXPROCS=1 go run ./05-falsesharing/stats|Two lines, Stats then PaddedStats. Stats 145-155 ms, PaddedStats 146-153 ms over 10 runs of the built binary: no difference, per-run Stats/PaddedStats 0.97-1.03 (median 1.0), Stats the slower in only 4 of 10|Если Stats и PaddedStats отличаются на несколько миллисекунд, это нормально: на одном CPU горутины работают по очереди, и делить кэш-линию не с кем. Запоминаем число около 150 мс, с ним сравниваем следующий запуск.
D5|2, 9.3|short|go run ./05-falsesharing/stats|Stats 296-344 ms, PaddedStats 72-74 ms over 10 runs of the built binary; per-run Stats/PaddedStats 4.0-4.8x (median 4.4), Stats slower in 10 of 10. Against D5p1 Stats doubles (median 315.5 against 151.0 ms, 2.1x) and PaddedStats halves (72.5 against 150.0 ms)|Если у нас вышло не в четыре раза, а в три или в пять, это нормально. Важно, что на двух CPU Stats медленнее, чем на одном, а PaddedStats быстрее. Если разницы нет, запускаем еще раз: куда попали две горутины, из WSL не видно.
D1|3|short|go run ./01-counter|want 800000; unsynchronized got far below 800000 (312717-530238 over 25 runs; the figure moves with machine load), never 800000; sync.Mutex, atomic.Int64 and channel-owned got 800000 lost 0. Wall 0.1 s|Вышло ровно 800 000? Это и есть урок: правильный ответ из программы с гонкой, и он ничего не доказывает.
D1p1|3|short|GOMAXPROCS=1 go run ./01-counter|want 800000; unsynchronized got 800000 lost 0 on one CPU: 25 of 25 runs; the same program is 0 of 25 at the default 20. Under GOMAXPROCS=1 go run -race the race is still reported, 5 of 5. Wall 0.08 s|На одном ядре горутины обычно не пересекаются, но никто этого не обещает. Значит, и здесь ответ может выйти неправильным.
D1p1r|3|short|GOMAXPROCS=1 go run -race ./01-counter|want 800000; one WARNING: DATA RACE at main.go:21 (read by a running goroutine, previous write by one shown as (finished)); unsynchronized got 800000 lost 0; Found 1 data race(s); exit status 66, and go run itself exits 1. Reported and right in 5 of 5 runs. Wall 1.3 s, one second of it the race runtime exit sleep|Отчета нет? Запускаем еще раз. До сих пор он был в каждом запуске, и ответ при этом был правильный.
D1r|3|-|go run -race ./01-counter|Two WARNING: DATA RACE reports at main.go:21 (read and write, then write and write), created at sync.(*WaitGroup).Go; Found 2 data race(s); exit status 66, and go run itself exits 1. Wall 1.5 s, one second of it the race runtime exit sleep|Номера горутин и адрес в отчете каждый раз другие, это нормально. Смотрим на строку main.go:21 и на exit status 66 в конце.
D4-01|5.3|-|go run -race ./04-edges 01|=== 01 go statement; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-02|5.3|-|go run -race ./04-edges 02|=== 02 goroutine exit <-- expect race; one DATA RACE report; exit status 66. Wall 1.05 s|Отчета нет? Запускаем еще раз. В пяти запусках из пяти отчет был: Sleep разносит обращения по времени, но ребра happens-before не дает.
D4-03|5.3|-|go run -race ./04-edges 03|=== 03 chan send -> recv; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-04|5.3|short|go run -race ./04-edges 04|=== 04 unbuffered recv -> send; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-05|5.3|short|go run -race ./04-edges 05|=== 05 cap-1 recv -> send <-- expect race; one DATA RACE report, write at main.go:71, previous read at main.go:73; exit status 66. Wall 1.05 s|Отчета нет? Запускаем еще раз. В пяти запусках из пяти он был: от 04 этот случай отличается только емкостью канала.
D4-06|5.3|-|go run -race ./04-edges 06|=== 06 close -> recv; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-07|5.3|-|go run -race ./04-edges 07|=== 07 Unlock -> Lock; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-08|5.3|-|go run -race ./04-edges 08|=== 08 sync.Once; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-09|5.3|-|go run -race ./04-edges 09|=== 09 atomic Store -> Load; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D4-10|5.3|-|go run -race ./04-edges 10|=== 10 plain bool flag <-- expect race; two DATA RACE reports (the flag and the payload); exit status 66. Wall 1.05 s|Отчет один, а не два? Это тоже гонка, и смысл тот же: у plain bool нет ребра happens-before.
D4-11|5.3|-|go run -race ./04-edges 11|=== 11 WaitGroup.Go / Wait; no report; exit 0. Wall 1.05 s|Если чистый случай вдруг дал отчет, сначала смотрим, какая переменная в отчете. Если это не payload, ошибка в самом демо, а не в модели памяти.
D8|6.3|short|go run ./08-tearing|part 1: 51-54% of struct reads observed a != b; part 2: 38-42% of string reads had a length from one string and a data pointer from the other; exit 0. Read counts vary with the core. Wall 0.6 s|Если доля другая, это от скорости ядер. Важно, что она не ноль: мы читаем значение, которое никто не записывал.
D2a|7.2|-|sh 02-compiler/run.sh|Three listings: SpinPlain size=10 (CMPB, JEQ 0, RET), SpinAtomic size=12 (XCHGL AX, AX at spin.go:22, then MOVL, TESTL, JEQ 0), SumShared size=23 (ADDQ main.shared(SB) inside the loop). Wall 0.1 s|Если размеры другие, значит другая версия Go или другая архитектура: листинг снят на go1.27.1 под amd64. Главное не меняется: atomic load здесь тот же MOVL.
D2|7.2|-|go run ./02-compiler|plain-bool spin loop: terminated; atomic.Bool spin loop: terminated. Wall 0.05 s|Если первая строка STILL SPINNING, это ровно то, что модель памяти разрешает: цикл не обязан увидеть запись.
D3|7.3|short|go run ./03-reorder|rounds=1000000 each; plain int32 seen 6297-40156 times (0.6-4%), a different count every run; atomic.Int32 seen 0 times, every run. Wall 0.5 s|Число каждый раз другое, разброс в разы, и это нормально. Если plain дал ноль, просто запускаем еще раз; важно, что atomic всегда ноль.
D6|8.3, 9.2|-|go test -run=X -bench=. -count=5 -cpu=1,2 ./06-atomiccost|100 benchmark lines in about 100 s: start it, then read the source. PlainLoad and AtomicLoad 0.18-0.19 ns/op, PlainStore 0.15-0.18, AtomicStore 3.45-3.54, Add 3.65-3.76, CAS 3.93-4.13, Lock+Unlock 10.0-10.4; AtomicLoadShared halves at -2, AtomicAddShared 6.1-6.3 then 7.9-8.6. Wall 97-100 s|На другой машине наносекунды будут другие. Смотрим на две вещи: atomic load стоит как обычный load, а store и CAS на порядок дороже.
D7|10.3|short|go run ./07-aba|4 workers, 20000 ops each; three rows stack intact; recycled nodes, widened window CORRUPT: 6 or 7 nodes lost, cycle=true, every run. Wall 0.1 s|Сколько узлов потеряно, каждый раз по-разному. Важна форма: ломается только последняя строка, три остальные целы.'

if [ "$short" = 1 ]; then
	CUES=$(printf '%s\n' "$CUES" | grep '^[^|]*|[^|]*|short|')
fi
if [ -n "$only" ]; then
	CUES=$(printf '%s\n' "$CUES" | grep "^$only|" || :)
	if [ -z "$CUES" ]; then
		echo "cues.sh: no cue $only; sh cues.sh --list shows the ids." >&2
		exit 2
	fi
fi
ncues=$(printf '%s\n' "$CUES" | wc -l | tr -d ' ')

# show N: print cue N (a line number in $CUES) and set id, kind, cmd, expect, fallback.
show() {
	IFS='|' read -r id sect tag cmd expect fallback <<-EOF
	$(printf '%s\n' "$CUES" | sed -n "${1}p")
	EOF
	if [ "$present" = 1 ] && [ "$list" = 0 ]; then
		printf '\033[2J\033[H'
	else
		echo
	fi
	echo "== cue $id   [§$sect]   $1 of $ncues"
	echo "   \$ $cmd"
	[ "$present" = 1 ] && [ "$list" = 0 ] && return 0
	echo "   Expect: $expect"
	[ "$list" = 1 ] && echo "   Если не так: $fallback"
	return 0
}

if [ "$list" = 1 ]; then
	echo "Cues for L-03, from demos.md \"Run order for the lecture\". Run from this directory."
	echo "Expected ranges: $STAMP."
	i=1
	while [ "$i" -le "$ncues" ]; do
		show "$i"
		i=$((i + 1))
	done
	exit 0
fi

# The pin is enforced by the go directive in go.mod, not by the command name.
# A shell that exports GOROOT for an older Go (say /usr/local/go, go1.25.3)
# breaks it: with the go1.27.1 binary first on PATH that GOROOT wins, go env
# GOROOT, go list and go doc read the older tree, and a build fails with
#   compile: version "go1.25.3" does not match go tool version "go1.27.1"
# Drop a GOROOT that is not the pinned tree.
if [ -n "${GOROOT:-}" ] && [ "$(head -1 "$GOROOT/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "cues.sh: your shell exports GOROOT=$GOROOT, not the go1.27.1 tree; unsetting it for this run."
	unset GOROOT
fi
v=$(go env GOVERSION)
if [ "$v" != "go1.27.1" ]; then
	echo "cues.sh: go here is $v, want go1.27.1." >&2
	echo "cues.sh: check GOTOOLCHAIN (want auto) and that this is inside the course repo." >&2
	exit 1
fi
G=$(go env GOROOT)
if [ "$(head -1 "$G/VERSION" 2>/dev/null)" != "go1.27.1" ]; then
	echo "cues.sh: go env GOROOT is $G, whose VERSION is not go1.27.1." >&2
	exit 1
fi

# A GODEBUG, GOMAXPROCS or GORACE left in the shell would leak into every cue
# (GORACE also changes the race runtime's one-second exit sleep).
for e in GODEBUG GOMAXPROCS GORACE; do
	if eval "[ -n \"\${$e+x}\" ]"; then
		echo "cues.sh: unsetting $e, which your shell exports; the cues set their own."
		unset "$e"
	fi
done

# runcue: run $cmd and report a non-zero status. A child of a shell that
# ignores INT inherits SIG_IGN, so INT gets a handler (trap :), not trap ''.
runcue() {
	rc=0
	eval "$cmd" || rc=$?
	case $rc in
	0) ;;
	130) echo "== stopped by Ctrl-C (exit status 130)" ;;
	*) echo "== exit status $rc" ;;
	esac
	if [ "$present" = 0 ]; then
		echo "   Если не так: $fallback"
	fi
}

if [ -n "$only" ]; then
	trap : INT
	show 1
	runcue
	exit 0
fi

if [ ! -t 0 ]; then
	echo "cues.sh: needs a terminal for the key prompt; use --list to print the cues." >&2
	exit 1
fi

tmp=$(mktemp -d "${TMPDIR:-/tmp}/cues.XXXXXX")
saved=$(stty -g)
trap 'stty "$saved" 2>/dev/null; rm -rf "$tmp"' EXIT
trap 'exit 129' HUP
trap 'exit 143' TERM

# Warm the build cache, so the first go run, go run -race and go test are not
# slow on screen. vet compiles every package, test-only ones included; the
# -race builds and the test links are separate cache entries.
echo "== warming the build cache with $v, $(go env GOOS)/$(go env GOARCH)"
if ! {
	go vet ./... &&
		go build -o "$tmp/" ./01-counter ./02-compiler ./03-reorder ./04-edges ./05-falsesharing/stats ./07-aba ./08-tearing &&
		go build -race -o "$tmp/race-01" ./01-counter &&
		go build -race -o "$tmp/race-04" ./04-edges &&
		go test -c -o "$tmp/05.test" ./05-falsesharing &&
		go test -c -o "$tmp/06.test" ./06-atomiccost
} >"$tmp/warm.log" 2>&1; then
	cat "$tmp/warm.log" >&2
	echo "cues.sh: the warm-up failed." >&2
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

trap : INT

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
	runcue
	i=$((i + 1))
done
