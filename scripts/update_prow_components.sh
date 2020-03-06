#!/usr/bin/env bash

root_dir="$(dirname "$(realpath "$0")")"
source "$root_dir"/preflight_checks.sh

new_tag="$1"
if [[ -z $new_tag ]]; then
  echo "missing required new tag argument" >&2
  exit 1
fi

echo "Updating pod utilities (excluding clonerefs): to $new_tag"

for i in initupload sidecar entrypoint; do
  find $root_dir/../config -type f -exec $SED -i "s/(.*${i}:)[^\"]+/\1$new_tag/g" {} +
done

echo "Updating checkconfig to $new_tag"

for path in ../config boilerplate; do
  find $root_dir/$path -type f -exec $SED -i "s/(.*checkconfig:)[^\"]+/\1$new_tag/g" {} +
done

$SED -i "s/(.*checkconfig:)[^\ ]+/\1$new_tag/g" Makefile

echo
echo "Please verify changes are as intended"
git status --short
