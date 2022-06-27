#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")" || exit
mkdir -p out
docker run \
  --rm \
  -v "$(pwd)":/project \
  -v "$(pwd)/out":/go \
  golang:1-alpine \
  sh -c "cd /project && go install && chown -R $(id -u "${USER}"):$(id -g "${USER}") /go/bin"
mv out/bin/* ./
