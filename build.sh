#!/bin/bash

GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) go build -o oauth-proxy
strip oauth-proxy
