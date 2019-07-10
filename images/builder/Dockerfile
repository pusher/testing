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

#######################################################################
###
### GOLANG BUILDER - For building any go binaries needed in later steps
###
#######################################################################

FROM golang:1.12 AS builder
COPY barnacle/* $GOPATH/src/barnacle/

RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR $GOPATH/src/barnacle
RUN dep ensure --vendor-only
RUN go build -o /barnacle $GOPATH/src/barnacle

# Add the AWS credential helper
RUN go get -u github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cli/docker-credential-ecr-login

##################################################
###
### Builder image including Docker-In-Docker setup
###
##################################################

FROM debian:stretch

WORKDIR /workspace
RUN mkdir -p /workspace
ENV WORKSPACE=/workspace \
    TERM=xterm

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

# Add the testing repository so we can get newer Python versions (> 3.5)
RUN echo 'deb http://ftp.de.debian.org/debian testing main' >> /etc/apt/sources.list

# common util tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    file \
    git \
    jq \
    mercurial \
    openssh-client \
    pkg-config \
    procps \
    python \
    python3 \
    python-pip \
    python3-pip \
    python-setuptools \
    python3-setuptools \
    rsync \
    unzip \
    wget \
    xz-utils \
    zip \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

#
# BEGIN: DOCKER IN DOCKER SETUP
#

# Install Docker deps, some of these are already installed in the image but
# that's fine since they won't re-install and we can reuse the code below
# for another image someday.
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    lsb-release && \
    rm -rf /var/lib/apt/lists/*

# Add the Docker apt-repository
RUN curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg \
    | apt-key add - && \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
    $(lsb_release -cs) stable"

# Install Docker
RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce=18.06.* && \
    rm -rf /var/lib/apt/lists/* && \
    sed -i 's/cgroupfs_mount$/#cgroupfs_mount\n/' /etc/init.d/docker


# Move Docker's storage location
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --data-root=/docker-graph"' | \
    tee --append /etc/default/docker
# NOTE this should be mounted and persisted as a volume ideally (!)
# We will make a fallback one now just in case
RUN mkdir /docker-graph

# add custom docker cleanup binary
COPY --from=builder ["/barnacle", "/usr/local/bin/"]

# add ecr login help
COPY --from=builder ["/go/bin/docker-credential-ecr-login", "/usr/local/bin/"]

# Install docker-compose
RUN pip install docker-compose

#
# END: DOCKER IN DOCKER SETUP
#

#
# BEGIN: SNYK SECURITY SCANNER SETUP
#

# install nodejs
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && curl -sL https://deb.nodesource.com/setup_12.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# install the snyk binary
RUN npm install -g snyk

#
# END: SNYK SECURITY SCANNER SETUP
#

# Run everything from here on out as the prow user
RUN apt-get update && \
    apt-get install -y --no-install-recommends sudo && \
    rm -rf /var/lib/apt/lists/*
RUN echo "prow ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
RUN useradd -rm -d /home/prow -s /bin/bash -g root -G sudo,docker -u 1000 prow
RUN chown prow:root /workspace \
    && chmod g+s /workspace
RUN mkdir -p /home/prow/go \
    && chown prow:root /home/prow/go \
    && chmod g+s /home/prow/go
USER prow

# note the runner is responsible for making docker in docker function if
# env DOCKER_IN_DOCKER_ENABLED is set
COPY ["runner","/usr/local/bin/"]

ENTRYPOINT ["/usr/local/bin/runner"]
