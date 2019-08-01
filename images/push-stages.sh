#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

repository="quay.io/pusher"

for arg in "$@"; do
  case ${arg%%=*} in
   "--build-root")
      build_root="${arg##*=}"
      ;;
    "--help")
      printf "${GREEN}$0${NC}\n"
      printf "  available options:\n"
      printf "  --build-root=${BLUE}<docker build root>${NC}\n"
      exit 0
      ;;
    *)
      echo "Unknown option: $arg"
      exit 2
      ;;
    esac
done


###########################################
#
# Check if the image is a multi stage build
#
###########################################
lines=$(cat ${build_root}/Dockerfile | grep 'FROM')
declare -a stages
while read -r from; do
  # Skip FROMs that aren't named
  if [[ ! $from =~ "AS" ]]; then
    continue
  fi
  stage=$(echo $from | sed -E 's|FROM .* AS (.*)$|\1|')
  stages+=($stage)
done <<< "$lines"

#####################
#
# Push any pre-stages
#
#####################
for stage in "${stages[@]}"; do
  echo -e "${CYAN}Pushing Docker pre-stage: ${build_root}:${stage}${NC}"
  docker push ${repository}/${build_root}:stage-${stage}
done
