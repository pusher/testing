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

ifeq ($(REPOSITORY),)
REPOSITORY := docker.io/pusher
endif

# Set the version parameter if it is unset or empty
ifeq ($(VERSION),)
VERSION := $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)
endif

all: build

####
#
# The following targets are for building all images
#
####

build: builder golang-builder kubebuilder-builder clonerefs python-builder ruby-builder dhall-builder admin-builder yamllint

tag: docker-tag-builder docker-tag-golang-builder docker-tag-kubebuilder-builder docker-tag-clonerefs docker-tag-python-builder docker-tag-ruby-builder docker-tag-dhall-builder docker-tag-admin-builder docker-tag-yamllint

push: docker-push-builder docker-push-golang-builder docker-push-kubebuilder-builder docker-push-clonerefs docker-push-python-builder docker-push-ruby-builder docker-push-dhall-builder docker-push-admin-builder docker-push-yamllint

####
#
# Targets for building specific images
#
####

builder: docker-build-builder

golang-builder: builder docker-build-golang-builder

kubebuilder-builder: golang-builder docker-build-kubebuilder-builder

python-builder: builder docker-build-python-builder

ruby-builder: builder docker-build-ruby-builder

dhall-builder: builder docker-build-dhall-builder

admin-builder: builder docker-build-admin-builder

clonerefs: docker-build-clonerefs

yamllint: docker-build-yamllint

####
#
# Generic docker related targets
#
###

docker-new-%:
	@mkdir -p $*
	@sed 's|%NAME%|$*|g' Dockerfile.boilerplate > $*/Dockerfile
	@echo -e "\033[36mA boilerplate Dockerfile has been generated at images/$*/Dockerfile\033[0m"

docker-build-%:
	./docker-build.sh --repository=$(REPOSITORY) --version=$(VERSION) --build-root=$*
	@echo -e "\033[36mBuilt $(REPOSITORY)/$*:$(VERSION)\033[0m"

# Set the tags parameter if it is unset or empty
ifeq ($(TAGS),)
TAGS := ${VERSION},latest
endif
docker-push-%: docker-build-% docker-tag-%
	./push-stages.sh --repository=$(REPOSITORY) --build-root=$*
	@echo -e "\033[36mPushing $(REPOSITORY)/$*:$(VERSION)\033[0m"
	@IFS=","; tags=${TAGS}; for tag in $${tags}; do docker push $(REPOSITORY)/$*:$${tag} && echo -e "\033[36mPushed $(REPOSITORY)/$*:$${tag}\033[0m"; done

docker-tag-%: docker-build-%
	@IFS=","; tags=${TAGS}; for tag in $${tags}; do docker tag  $(REPOSITORY)/$*:$(VERSION) $(REPOSITORY)/$*:$${tag} && echo -e "\033[36mTagged $(REPOSITORY)/$*:$(VERSION) as $${tag}\033[0m"; done
