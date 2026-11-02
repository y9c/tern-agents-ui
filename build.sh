#!/bin/sh
# Builds tclaude, tcodex, topencode and tpi into bin/ next to this script,
# plus thin passthrough shims named like the real CLIs (opencode, codex,
# claude, pi) that exec the installed tool, so a wrapper always uses the
# latest version and behaves exactly like it.
set -eu
cd "$(dirname "$0")"
go build -o bin/tclaude ./cmd/tclaude
go build -o bin/tcodex ./cmd/tcodex
go build -o bin/topencode ./cmd/topencode
go build -o bin/tpi ./cmd/tpi
for t in opencode codex claude pi; do
	go build -o "bin/$t" ./cmd/pass
done
