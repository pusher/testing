# Images

This folder contains Dockerfiles for building "builder" images to be used within
ProwJobs to execute CI tasks.

## Builder (base image)

The builder image (`quay.io/pusher/builder`) acts as a base image for all other
builder images.

The builder image contains the following:
- Docker-In-Docker installation
- A `prow` user and workspace configured with correct permissions
- A `runner` script that optionally starts Docker-In-Docker before executing
its `args` inside a bash shell

## Clonerefs

Clonerefs is **not** a builder image. Clonerefs is part of the Prow pod utilities
which clones code before ProwJobs start.
We maintain a copy to set the user to `prow` so that cloned repositories
have the correct permissions set.

## Golang Builder

The Golang builder (`quay.io/pusher/golang-builder`) is for building Golang
based projects within Prow.

The Golang builder contains the following:
- Everything in the base builder image
- Golang 1.13.5
- [profile](github.com/pkg/profile)
- [delve](github.com/go-delve/delve)
- [dep](github.com/golang/dep)
- [golangci-lint](github.com/golangci/golangci-lint)
- [ginkgo](github.com/onsi/ginkgo)

## Kubebuilder Builder

The Kubebuilder builder (`quay.io/pusher/golang-builder`) is for building
Kubebuilder based projects within Prow.

The Kubebuilder builder contains the following:
- Everything in the Golang builder image
- Kubebuilder testing tools for Kubernetes 1.11, 1.12 and 1.13
  - etcd
  - kube-apiserver
  - kubectl
  - kubebuilder

## Python Builder

The Python builder (`quay.io/pusher/python-builder`) is for testing Python based
projects within Prow.

The Python builder contains the following:
- Everything in the base builder image
- python 2.7.13 (under `python2`)
- Python 3.7.3 (under `python`)
- Pip 19.1
