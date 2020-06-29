#!/usr/bin/env bash

set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
# LINTER="$("$DIR/gobin.sh" -p github.com/golangci/golangci-lint/cmd/golangci-lint)"
SHELLFMTPATH="$("$DIR/gobin.sh" -p mvdan.cc/sh/v3/cmd/shfmt)"

# shellcheck source=./lib/logging.sh
source "$DIR/lib/logging.sh"

if [[ -n $CI ]]; then
	TEST_TAGS=${TEST_TAGS:-tm_test,tm_int}
else
	TEST_TAGS=${TEST_TAGS:-tm_test}
fi

if [[ $TEST_TAGS == *"tm_int"* ]]; then
	BENCH_FLAGS=${BENCH_FLAGS:--bench=^Bench -benchtime=1x}
fi

# Run shellcheck on shell-scripts, only if installed.
if command -v shellcheck >/dev/null 2>&1; then
	info "Running shellcheck"
	# shellcheck disable=SC2038
	if ! find . -name '*.sh' | xargs -n1 shellcheck -P SCRIPTDIR -s bash; then
		error "shellcheck failed on some files"
		exit 1
	fi
else
	echo "Warn: Not running shellscript linter due to shellcheck not being installed" >&2
fi

info "Running shfmt"
if ! "$SHELLFMTPATH" -s -d "$DIR/../"; then
	error "shfmt failed on some files"
	exit 1
fi

# TODO(jaredallard): enable golangci-lint
# "$LINTER" -c "$(dirname "$0")/golangci.yml" --build-tags "$TEST_TAGS" run ./...

info "Running go test ($TEST_TAGS)"
# Why: We want these to split.
# shellcheck disable=SC2086
go test $BENCH_FLAGS \
	-ldflags "-X github.com/tritonmedia/pkg/app.Version=testing" -tags="$TEST_TAGS" \
	-covermode=atomic -coverprofile=/tmp/coverage.out -cover "$@" ./...
