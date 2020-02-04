# Images

This folder contains Dockerfiles for building "builder" images to be used within
ProwJobs to execute CI tasks.

If none of the existing images suite your needs see "How to add a new builder image"
for how to create a new one.

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

## Ruby Builder

The Ruby builder (`quay.io/pusher/ruby-builder`) is for testing Ruby based
projects within Prow.

The Ruby builder contains the following:
- Everything in the base builder image
- ruby 2.6.5

## How to add a new \*-builder image

1. Run `make docker-new-$NAME`.
2. Add your new targets to the Makefile to build/tag/push your new image, following
   the existing conventions.
3. Add your modifications to the Dockerfile that the previous command generated.
4. Build the (base) builder image and push it to `quay.io`:
```
$ docker login --username <USERNAME> quay.io
# docker-push-* implicitly builds, tags and pushes
$ make docker-push-builder`.
```
5. Build your image `make docker-build-$NAME`.
6. When you're happy with the result, commit your changes, push to Github and open a PR.
> Note: due to how caching works in the build process you will need to rebuild and push
> the `builder` image after each commit, so (for now) it's recommended to make commits
> _after_ you're done with your modifications.
