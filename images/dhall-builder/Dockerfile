# Copyright 2020 Pusher Ltd.
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

#############################################################################
###
### DHALL BUILDER - FROM pusher/builder with Dhall tooling
###
#############################################################################

ARG VERSION=latest
ARG REPOSITORY=docker.io/pusher
FROM ${REPOSITORY}/builder:${VERSION}

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

ARG DHALL_VERSION=1.34.0

# Switch to root user to perform installation
USER root

RUN apt-get install -yq jq curl wget

RUN curl -sL https://api.github.com/repos/dhall-lang/dhall-haskell/releases/tags/${DHALL_VERSION} | jq -r '.assets[].browser_download_url' | wget -qi -  \
    && ls -1 dhall-*-x86_64-linux.tar.bz2 | xargs -n1 tar --extract --bzip2 -C /usr/local/bin --strip-components=2 -f \
    && rm -rf *.tar.bz2 *.zip

# Pre-cache commonly used Dhall packages
COPY ["pre-cache", "/usr/local/bin/dhall-pre-cache"]

RUN /usr/local/bin/dhall-pre-cache

# Switch back to prow user
USER prow
