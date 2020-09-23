#!/usr/bin/env bash

set -e

declare -a repositories=(quay.io/pusher docker.io/pusher)

for repository in "${repositories[@]}"; do
  env REPOSITORY="$repository" make --directory=images push
done
