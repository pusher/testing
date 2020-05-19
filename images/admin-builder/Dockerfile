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
### ADMIN BUILDER - FROM pusher/builder with Kubectl, AWS, GCLOUD and
###                 Terraform CLIs
###
#############################################################################

ARG VERSION=latest
ARG REPOSITORY=docker.io/pusher
FROM ${REPOSITORY}/builder:${VERSION}

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

# Switch to root user to perform installation
USER root

# Install Kubectl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
		&& chmod +x ./kubectl \
		&& mv ./kubectl /usr/local/bin/kubectl

# Install AWS CLI & check it was installed
RUN pip3 install awscli --upgrade \
    && aws --version

# Install GCLOUD CLI & check it was installed
RUN curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-250.0.0-linux-x86_64.tar.gz \
    && tar -zxvf google-cloud-sdk-250.0.0-linux-x86_64.tar.gz google-cloud-sdk \
		&& google-cloud-sdk/install.sh
ENV PATH=/workspace/google-cloud-sdk/bin:${PATH}
RUN gcloud --version

# Install Terraform CLI & check it was installed
ARG TF_VERSION="0.11.14"
RUN curl https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_amd64.zip -o terraform_${TF_VERSION}_linux_amd64.zip \
 && unzip terraform_${TF_VERSION}_linux_amd64.zip -d /usr/local/bin \
 && terraform --version

# Switch back to prow user for publish
USER prow
