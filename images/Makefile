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

REPO := quay.io/pusher
TAG ?= $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)

all: build

build: docker-build-builder

tag-latest: docker-tag-latest-builder

push: docker-push-builder

push-latest: docker-pusher-latest-builder

docker-build-%:
	docker build --pull --build-arg IMAGE_ARG=$(REPO)/$*:$(TAG) -t $(REPO)/$*:$(TAG) $*
	@echo "\033[36mBuilt $(REPO)/$*:$(TAG)\033[0m"

docker-push-%: docker-build-%
	docker push $(REPO)/$*:$(TAG)
	@echo "\033[36mPushed $(REPO)/$* with tag :$(TAG)\033[0m"

docker-tag-latest-%: docker-build-%
	docker tag $(REPO)/$*:$(TAG) $(IMG):latest
	@echo "\033[36mTagged $(REPO)/$*:$(TAG) with latest\033[0m"

docker-push-latest-%: docker-build-% docker-tag-latest-%
	docker push $(REPO)/$*:latest
	@echo "\033[36mPushed $(REPO)/$* with tag :latest\033[0m"
