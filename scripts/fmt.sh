#!/bin/sh

for pkg in $(go list -f {{.Dir}} ./... | grep -v /vendor/); do
  go fmt $pkg/*.go
done
