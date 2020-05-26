#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

repository="docker.io/pusher"

for arg in "$@"; do
  case ${arg%%=*} in
    "--version")
      version="${arg##*=}"
      ;;
    "--repository")
      repository="${arg##*=}"
      ;;
   "--build-root")
      build_root="${arg##*=}"
      ;;
    "--help")
      printf "${GREEN}$0${NC}\n"
      printf "  available options:\n"
      printf "  --version=${BLUE}<image version tag>${NC}\n"
      printf "  --build-root=${BLUE}<docker build root>${NC}\n"
      exit 0
      ;;
    *)
      echo "Unknown option: $arg"
      exit 2
      ;;
    esac
done

##############################################################################
#
# Use build cache from the given version if it exists, else use the latest tag
#
##############################################################################
curl --location --fail --show-error --silent --header "Accept: application/vnd.docker.distribution.manifest.v2+json" \
"https://quay.io/v2/pusher/$build_root/manifests/$version" > /dev/null
if [ $? != 0 ]; then
  # Tag does not exist, cache from latest
  cache_tag="latest"
else
  # Tag exists, cahce from this version
  cache_tag=$version
fi

###########################################
#
# Check if the image is a multi stage build
#
###########################################
lines=$(grep 'FROM .* AS' ${build_root}/Dockerfile)
declare -a stages
while read -r from; do
  stage=$(echo $from | sed -E 's|FROM .* AS (.*)$|\1|')
  stages+=($stage)
done <<< "$lines"

for stage in "${stages[@]}"; do
  echo -e "${CYAN}Building Docker pre-stage: ${build_root}:${stage}${NC}"
  docker pull ${repository}/${build_root}:stage-${stage}
  img=$repository/${build_root}
  docker build --pull --build-arg IMAGE_ARG=${img}:${version} --build-arg VERSION=${version} --build-arg REPOSITORY=${repository} --cache-from ${img}:stage-${stage} -t ${img}:stage-${stage} --target ${stage} ${build_root}
  stage_cache_from+="--cache-from $img:stage-$stage "
done



# cache_tag should always exists so now we can pull it
docker pull $repository/${build_root}:$cache_tag

########################
#
# Build the docker image
#
########################
echo -e "${CYAN}Building Docker image: ${build_root}${NC}"
img=${repository}/${build_root}
docker build --pull --build-arg IMAGE_ARG=${img}:${version} --build-arg VERSION=${version} --build-arg REPOSITORY=${repository} --cache-from ${img}:${version} ${stage_cache_from} -t ${img}:${version} ${build_root}
