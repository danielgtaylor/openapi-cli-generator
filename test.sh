#!/bin/sh

set -ev

# Build `openapi-cli-generator`
go generate
go install

# Generate our test example app
cd example-cli
rm -rf main.go
openapi-cli-generator init example
openapi-cli-generator generate openapi.yaml
sed -i'' -e 's/\/\/ TODO: Add register commands here./openapiRegister(false, globalFlags)/' main.go
go install
cd ..

cat >$HOME/.example/settings.toml <<EOL
[auth_servers]
[auth_servers.default]
client_id = ""
issuer = ""

[profiles]
[profiles.default]
api_url = "https://www.test.sh"
EOL

# Run all the tests!
go test "$@" ./...
