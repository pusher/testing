#!/usr/bin/env bash

config=""
for f in $(ls config/jobs/*/*.yaml); do
  k=${f#"config/jobs/"}
  k=$(echo $k | sed "s|/|_|")
  config="$config --from-file=$k=$f"
done
kubectl create configmap jobs ${config} --namespace default --dry-run --output yaml
