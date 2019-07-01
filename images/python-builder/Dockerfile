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
### PYTHON BUILDER - From quay.io/pusher/builder with Pythong tooling for
###                  building Python based projects
###
#############################################################################

ARG VERSION=latest
FROM quay.io/pusher/builder:${VERSION}

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

# Switch to root user to perform installation
USER root

#################
###
### Install Python
###
#################

# Install Python
RUN apt-get update && apt-get install -y \
		python \
		python3 \
		python-pip \
		python3-pip \
		python-setuptools \
		python3-setuptools \
	&& rm -rf /var/lib/apt/lists/*

#####################################
###
### Install tools for python projects
###
#####################################

# Tools for linting
RUN pip install flake8 pylint
RUN pip3 install flake8 pylint


# Switch back to prow user for publish
USER prow
