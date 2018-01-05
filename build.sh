#!/usr/bin/env bash

echo "### Building ..."

go clean ./...

go build ./...

go test ./...

echo "### Done"