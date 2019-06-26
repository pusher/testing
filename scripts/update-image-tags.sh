#!/usr/bin/env bash

set -euo pipefail

BOLD="\033[1m"
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
CYAN="\033[36m"
NC="\033[0m"

# A tag name must be valid ASCII and may contain lowercase and uppercase letters,
# digits, underscores, periods and dashes.
image_regex='quay.io/pusher/(.*)builder:([a-zA-Z0-9\._-]+)$'
pinned_image_regex='quay.io/pusher/(.*)builder:([a-zA-Z0-9\._-]+)\s+(.+)$'

TAG="${1:-}"

if [[ -z $TAG ]]; then
  echo -e "${RED}Must supply desired tag as first argument${NC}"
  exit 1
fi

echo -e "${CYAN}${BOLD}Updating images to tag '$TAG'${NC}"

for f in $(ls config/jobs/*/*.yaml config/README.md); do
  echo -e "Updating $f"
  r="s|$image_regex|quay.io/pusher/\1builder:$TAG|g"
  # Hack to make Linux and Mac sed work in-place
  if [[ $(uname) == "Darwin" ]]; then
    sed -E -i '' $r $f
  else
    sed -E -i $r $f
  fi
  echo -e "${GREEN}Updated $f${NC}"
done

for f in $(ls config/jobs/*/*.yaml); do
  while read -r line; do
    if [[ ! -z $line ]]; then echo -e "${YELLOW}${BOLD}[WARNING] Pinned Image in '$f':${NC} ${line#'image: '}"; fi
  done <<< $(grep -E $pinned_image_regex $f)
done
