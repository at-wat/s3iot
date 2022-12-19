#!/bin/bash

set -eu

if [ $# -ne 1 ]
then
  echo "usage: $(basename $0) version" >&2
  exit 1
fi

version=$1

for submod in awss3v1 awss3v2 examples
do
  sed -i "s|\(\s*github.com/at-wat/s3iot\) v.*|\1 ${version}|" ${submod}/go.mod
  sed -i '/^github.com\/at-wat\/s3iot/d' ${submod}/go.sum
done

git diff
