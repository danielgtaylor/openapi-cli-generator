#!/bin/sh

set -e

# Build `openapi-cli-generator`
go generate
go install

# Generate our test example app
cd example-cli
rm -rf main.go
openapi-cli-generator init example
openapi-cli-generator generate openapi.yaml
sed -i'' -e 's/\/\/ TODO: Add register commands here./openapiRegister(false)/' main.go
go install
cd ..

# Run all the tests!
GOCACHE=off go test "$@" ./...
