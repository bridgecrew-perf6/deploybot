#!/usr/bin/env bash
mkdir -p out
docker run \
  --rm \
  -v "$(pwd)":/project \
  -v "$(pwd)/out":/go \
  golang:1-alpine \
  sh -c "cd /project && go install && chown -R $(id -u "${USER}"):$(id -g "${USER}") /go/bin"
mv out/bin/* ./
