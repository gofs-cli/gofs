#!/usr/bin/env bash

set -e
set -u
set -o pipefail
set -x

go tool templ generate
go tool sqlc generate
bun run build
bun run tailwind
go build -o ./tmp/main cmd/server/main.go
