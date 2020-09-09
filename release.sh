#!/bin/bash

version="$1"

git tag "$version"
docker image build -t "jacksgt/u9k:$version" .
docker image push "jacksgt/u9k:$version"
