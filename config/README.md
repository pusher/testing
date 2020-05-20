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
| preset-root-docker-credentials: "true" | Add Quay credentials to allow the `root` user of the container push images to the Pusher Quay organisation. **Note**: This is used by the Istio ProwJobs as they have been copied from the upstream repository and run their builds as `root`. This should not be used for any Pusher ProwJob as these should be run as the `prow` user. |
| preset-docker-credentials: "true" | Add Quay credentials to allow the `prow` user of the container push images to the Pusher Quay organisation. |
| preset-dind-enabled: "true" | Starts a Docker daemon within the ProwJob container to allow `docker`commands to be executed. Used for building `Dockerfiles` and pushing images. |
| preset-golang-junit: "true" | Intercept calls to `go test` and generate a JUnit XML file for nicer test output within Prow's UI |

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

## Job Builder Images

Since most jobs should use one of the builder images from the [images](../images)
folder, the image tag for these images should stay the same, eg:

```
quay.io/pusher/builder:v20190821-328974b
```

Image tags are currently checked in CI and will be enforced to the version in
the example above.

### Updating the image version

To update the image version, replace the `IMAGE` argument in the root level
[Makefile](../Makefile) with the desired version.

Then run `make update-image-tags config` from the root of the repository to
update all jobs using a builder image and update the generated configuration,
then commit the result.

```bash
$ sed -i '' -E 's|^IMAGE \?= (.*)|IMAGE \?= <NEW_VERSION>|' Makefile
$ make update-image-tags config
$ git add .
$ git commit -m "Update image tags to <NEW_VERSION>"
```

## Pinning an image version

If for any reason you need to pin a builder image to a previous build;
Add a comment after the image tag explaining why the image is pinned and the
version update enforcement will ignore this line.

```
quay.io/pusher/builder:pinned # Pinned as an example to disable updater.
```
