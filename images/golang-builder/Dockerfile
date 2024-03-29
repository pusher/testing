# Copyright 2019 Pusher Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG VERSION=latest
ARG REPOSITORY=docker.io/pusher

#######################################################################
###
### GOLANG Pre-BUILDER - For building any go binaries needed in later steps
###
#######################################################################

FROM golang:1.13 AS builder
COPY wrapper/* $GOPATH/src/wrapper/

ENV GO111MODULE=on
WORKDIR $GOPATH/src/wrapper
RUN go test
RUN go build -o /wrapper $GOPATH/src/wrapper


#############################################################################
###
### GOLANG BUILDER - From pusher/builder with Go tooling for building
###                  Golang based projects
###
#############################################################################

FROM ${REPOSITORY}/builder:${VERSION}

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

# Switch to root user to perform installation
USER root

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
  g++ \
  gcc \
  libc6-dev \
  make \
  pkg-config \
  && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.13.5

RUN set -eux; \
  \
  # Copy-paste the relevant SHA256 checksum for the image from here:
  # https://golang.org/dl/
  # this is currently the go1.13.5.linux-amd64.tar.gz SHA
  goRelArch='linux-amd64'; goRelSha256='512103d7ad296467814a6e3f635631bd35574cab3369a97a323c9a585ccaa569' ;\
  \
  url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
  wget -q -O go.tgz "$url"; \
  echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
  tar -C /usr/local -xzf go.tgz; \
  rm go.tgz; \
  \
  export PATH="/usr/local/go/bin:$PATH"; \
  go version

# Switch back to prow user for publish
USER prow

RUN sudo mkdir /go \
    && sudo chown prow:root /go \
    && chmod g+s /go

# Install tools globally but owned by prow (if they are installed in
# /home/prow/go they are overwritten by Prow's code mountpoint)
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin"

#################################
###
### Install tools for go projects
###
#################################

RUN go get github.com/pkg/profile
RUN go get github.com/go-delve/delve/cmd/dlv
# Install dep for dependencies
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
# Install golangci-lint for linting
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.22.2
# Install ginkgo for running tests
RUN GO111MODULE=on go get github.com/onsi/ginkgo/ginkgo@v1.14.0

# Install more lint and code generation tools
RUN go get golang.org/x/tools/cmd/goimports \
    && go get github.com/kisielk/errcheck \
    && go get github.com/gordonklaus/ineffassign \
    && GO111MODULE=on go get github.com/fzipp/gocyclo@v0.3.0 \
    && GO111MODULE=on go get honnef.co/go/tools/...@2020.1.6 \
    && go get golang.org/x/lint/golint \
    && go get gopkg.in/golang/mock.v1/gomock \
    && go get gopkg.in/golang/mock.v1/mockgen

# Copy the go wrapper to /home/prow/.local/bin/go/
COPY --from=builder /wrapper /home/prow/.local/go/bin/go
