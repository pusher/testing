# Testing

This repository contains the configuration and deployment files for
[prow.pusher.com](https://prow.pusher.com).

## Repository Structure

### config

The [config](config) folder contains raw Prow configuration.
This configuration defines not only the global behaviour of the Prow
installation but also the jobs for the repositories managed by this Prow
cluster.

**Note**: Whenever changes are made to this folder, you **must** run
`make config` from the repository root to update the generated configuration
within the [prow](prow) folder and check this in with your changes.

### images

The [images](images) folder contains Dockerfiles defining a number of base
images from which ProwJobs can be run.

The base level image is called [builder](images/builder) and handles setup
of generic ProwJob bootstrapping such as configuring Docker-in-Docker.
All other images should inherit from this base image.

### prow

The [prow](prow) folder contains the Kubernetes deployment resources for
the Prow cluster.
These resources are automatically deployed by a [Faros](https://github.com/pusher/faros)
instance running within the Prow cluster.

### scripts

The [scripts](scripts) folder contains bash scripts for managing files
within this repository.
