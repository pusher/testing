#/env/bin/bash

RED="\033[31m"
GREEN="\033[32m"
CYAN="\033[36;1m"
NC="\033[0m"

failed=false

# A tag name must be valid ASCII and may contain lowercase and uppercase letters,
# digits, underscores, periods and dashes.
image_regex='^(.+)/(.+)/(.+):([a-zA-Z0-9\._-]+)(.*)'

for f in $(ls config/jobs/*/*.yaml); do
  echo "${CYAN}Processing $f${NC}"
  while read -r image; do
    image=$(echo $image | sed 's|.*image:||')
    registry=$(echo $image | sed -E "s|$image_regex|\1|")
    repository=$(echo $image | sed -E "s|$image_regex|\2/\3|")
    tag=$(echo $image | sed -E "s|$image_regex|\4|")
    echo "  Checking tag '$tag' for image '$registry/$repository'..."
    curl --location --fail --silent --header "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    "http://$registry/v2/$repository/manifests/$tag" &> /dev/null
    if [ $? != 0 ]; then
      echo "  ${RED}Tag '$tag' for image '$registry/$repository' not found!${NC}"
      failed=true
    else
      echo "  ${GREEN}Valid tag '$tag' found for image '$registry/$repository'${NC}"
    fi
  done <<< $(cat $f | grep 'image: ')
done

if [ "$failed" = true ]; then
  echo "${RED}Invalid images found.${NC}"
  exit 1
fi
