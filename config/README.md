# Config

This folder contains the raw configuration for Prow.

**Note**: Whenever changes are made to this folder, you **must** run
`make config` from the repository root to update the generated configuration
within the [prow](prow) folder and check this in with your changes.

## [config.yaml](config.yaml)

The config file provides configuration for Prow components such as Plank
and Deck as well as a number of global presets available to any ProwJob
within the cluster.

### Existing Presets

| Label | Description |
| ----- | ----------- |
| preset-service-account: "true" | Add Google application credentials to the environment of the ProwJob container. |
| preset-root-quay-credentials: "true" | Add Quay credentials to allow the `root` user of the container push images to the Pusher Quay organisation. **Note**: This is used by the Istio ProwJobs as they have been copied from the upstream repository and run their builds as `root`. This should not be used for any Pusher ProwJob as these should be run as the `prow` user. |
| preset-quay-credentials: "true" | Add Quay credentials to allow the `prow` user of the container push images to the Pusher Quay organisation. |
| preset-dind-enabled: "true" | Starts a Docker daemon within the ProwJob container to allow `docker`commands to be executed. Used for building `Dockerfiles` and pushing images. |

## [plugins.yaml](plugins.yaml)

The plugins file defines which [plugins](https://prow.pusher.com/plugins)
should be enabled for each particular repository managed by Prow.

**Note**: When adding a new repository to Prow you will need to add a
plugin config to this file.

## [jobs](jobs)

The [jobs](jobs) folder contains presubmit and postsubmit job definitions for
repositories managed by Prow.

**Note**: All files within this folder and its subfolders must be uniquely named
due to a limitation in how Prow consumes the configuration files.
