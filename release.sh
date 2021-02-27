#!/bin/bash

set -e

version="$1"

if echo "$version" | grep -E '^v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+$'; then
    echo "Releasing version '$version'"
else
    echo "Invalid release tag '$version'"
    exit 1
fi

if [ $(git diff --cached | wc -l) != '0' ]; then
    echo "You have staged changes! Aborting release"
    exit 1
fi

sed -i "s/const Version = \".*\"\$/const Version = \"$version\"/" config/config.go
git add config/config.go
git commit -m "Release version $version"
git tag "$version"
