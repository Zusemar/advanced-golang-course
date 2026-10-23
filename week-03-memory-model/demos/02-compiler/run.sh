#!/bin/sh
# Print each function's generated code.
#
# Run from the module root (the directory holding go.mod):
#     sh 02-compiler/run.sh
#
# The pinned toolchain must answer to `go`. If your shell exports GOROOT, run it as
#     env -u GOROOT sh 02-compiler/run.sh
# because a GOROOT that names another Go release can fail the build with a
# version mismatch. GO=<a go binary> overrides the command.
set -e
GO="${GO:-go}"

# -o is required: without it `go build` wants to write a binary named
# "02-compiler" over the directory of the same name and fails.
#
# -S prints every function in the package, main.main and its closures too, with
# absolute file paths. awk keeps the three functions' blocks, a blank line
# between them (a block ends at the next line that does not start with a tab),
# and drops the FUNCDATA and PCDATA bookkeeping lines; sed shortens the paths
# to spin.go and atomic/type.go.
"$GO" build -gcflags='-S' -o /dev/null ./02-compiler 2>&1 |
  awk '/^main\.(SpinPlain|SpinAtomic|SumShared) STEXT/ { if (n++) print ""; on = 1; print; next }
       !/^\t/ { on = 0 }
       on && /(spin|atomic\/type)\.go:/ && !/FUNCDATA|PCDATA/' |
  sed -e 's#(/[^()]*/\(spin\.go:\)#(\1#' -e 's#(/[^()]*/\(atomic/type\.go:\)#(\1#' -e 's/^\t//'
