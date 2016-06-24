#!/bin/bash

GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) go build -o oauth-proxy

if [ $? == 0 ]; then
    strip oauth-proxy
else
    exit 1
fi
