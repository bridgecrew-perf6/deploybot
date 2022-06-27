#!/usr/bin/env bash
mkdir -p out
docker run \
  --rm \
  -v "$(pwd)":/project \
  -v "$(pwd)/out":/go \
  golang:1.16-alpine \
  sh -c 'cd /project && go install'
mv out/bin/* ./
