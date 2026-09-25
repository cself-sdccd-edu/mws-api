#!/usr/bin/env bash
set -euo pipefail

VERSION="$(cat VERSION)"
COMMIT="$(git rev-parse --short HEAD)"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

mkdir -p dist

go fmt ./...
go vet ./...
#go test ./...

OUTPUT="dist/mws-api"

CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X 'github.com/cself-sdccd-edu/mws-api/internal/version.Version=${VERSION}' -X 'github.com/cself-sdccd-edu/mws-api/internal/version.Commit=${COMMIT}' -X 'github.com/cself-sdccd-edu/mws-api/internal/version.BuildDate=${BUILD_DATE}'" -o "${OUTPUT}" ./cmd/mws-api
