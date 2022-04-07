#!/bin/bash

set -eu

rm $(grep -rl 'Code generated by moq' | grep -v '/moq.yml$')
for mod in ./ ./awss3v1 ./awss3v2
do
  (
    cd ${mod}
    go mod tidy
    go generate ./...
  )
done
git restore $(find . -name 'go.mod') go.sum
rm ./awss3v1/go.sum ./awss3v2/go.sum
