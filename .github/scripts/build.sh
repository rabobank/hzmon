#!/bin/bash

OUTPUT_DIR=$PWD/dist
mkdir -p "${OUTPUT_DIR}"
echo "Building HzMon, VERSION=${VERSION}, COMMIT=${COMMIT}"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${OUTPUT_DIR}"/HzMon_linux_amd64 -ldflags "-X github.com/rabobank/hzmon/version.VERSION=${VERSION} -X github.com/rabobank/hzmon/version.COMMIT=${COMMIT}" .
