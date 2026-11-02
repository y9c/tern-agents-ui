#!/bin/sh
# Builds tclaude, tcodex and topencode into bin/ next to this script.
set -eu
cd "$(dirname "$0")"
go build -o bin/tclaude ./cmd/tclaude
go build -o bin/tcodex ./cmd/tcodex
go build -o bin/topencode ./cmd/topencode
