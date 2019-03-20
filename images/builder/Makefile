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

IMG = quay.io/pusher/builder
TAG ?= $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)

all: build

build:
	docker build --pull --build-arg IMAGE_ARG=$(IMG):$(TAG) -t $(IMG):$(TAG) .
	@echo "\033[36mBuilt $(IMG):$(TAG)\033[0m"

tag-latest:
	docker tag $(IMG):$(TAG) $(IMG):latest
	@echo "\033[36mTagged $(IMG):$(TAG) with latest\033[0m"

push: build
	docker push $(IMG):$(TAG)
	@echo "\033[36mPushed $(IMG) with tag :$(TAG)\033[0m"

push-latest: build tag-latest
	docker push $(IMG):latest
	@echo "\033[36mPushed $(IMG) with tag :latest\033[0m"
