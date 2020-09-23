#!/bin/bash

set -e

version="$1"

if [ $(git diff --cached | wc -l) != '0' ]; then
    echo "You have staged changes! Aborting release"
    exit 1
fi

sed -i "s/const Version = \".*\"\$/const Version = \"$version\"/" config/config.go
git add config/config.go
git commit -m "Release version $version"
git tag "$version"
docker image build -t "jacksgt/u9k:$version" .
docker image push "jacksgt/u9k:$version"
