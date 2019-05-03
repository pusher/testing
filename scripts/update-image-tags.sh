#/env/bin/bash

set -euo pipefail

RED="\033[31m"
GREEN="\033[32m"
CYAN="\033[36;1m"
NC="\033[0m"

# A tag name must be valid ASCII and may contain lowercase and uppercase letters,
# digits, underscores, periods and dashes.
image_regex='quay.io/pusher/(.*)builder:([a-zA-Z0-9\._-]+)$'

TAG="${1:-}"

if [[ -z $TAG ]]; then
  echo "${RED}Must supply desired tag as first argument${NC}"
  exit 1
fi

echo "${CYAN}Updating images to tag '$TAG'${NC}"

for f in $(ls config/jobs/*/*.yaml); do
  echo "Updating $f"
  sed -i '' -E "s|$image_regex|quay.io/pusher/\1builder:$TAG|g" $f
  echo "${GREEN}Updated $f${NC}"
done
