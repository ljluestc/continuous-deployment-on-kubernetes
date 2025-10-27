#!/usr/bin/env bash
set -euo pipefail

run_go_tests() {
	echo "Running Go tests in $1"
	pushd "$1" >/dev/null
	if [ -f go.mod ]; then
		go test ./... -cover -coverprofile=coverage.out
	else
		echo "Skipping $1 (no go.mod)"
	fi
	popd >/dev/null
}

# algorithm-visualization
run_go_tests "algorithm-visualization"

# sample-app
run_go_tests "sample-app"

# services/* modules with go.mod
for mod in services/*; do
	[ -d "$mod" ] || continue
	if [ -f "$mod/go.mod" ]; then
		run_go_tests "$mod"
	fi
done
