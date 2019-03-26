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

#############################################################################
###
### Kubebuilder BUILDER - From quay.io/pusher/golang-builder with Kubebuilder
###                       tooling for building kubebuilder projects
###
#############################################################################

ARG VERSION=latest
FROM quay.io/pusher/golang-builder:${VERSION}

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

# Install kubebuilder project tools (some of these may already exist)
ADD Makefile .

RUN sudo mkdir -p \
		/usr/local/kubebuilder-1.11/bin \
		/usr/local/kubebuilder-1.12/bin \
		/usr/local/kubebuilder-1.13/bin \
		&& sudo chown prow:root \
		/usr/local/kubebuilder-1.11/bin \
		/usr/local/kubebuilder-1.12/bin \
		/usr/local/kubebuilder-1.13/bin
RUN make \
		dep \
		ginkgo-cli \
		golangci-lint \
		kustomize \
		kubebuilder-tools-1.11 \
		kubebuilder-tools-1.12 \
		kubebuilder-tools-1.13 \
		&& rm Makefile
