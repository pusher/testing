BINARY := promoter
# Set the version parameter if it is unset or empty
ifeq ($(VERSION),)
VERSION := $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)
endif

# Image URL to use all building/pushing image targets
IMG ?= quay.io/pusher/promoter

.PHONY: all
all: test build

.PHONY: build
build: clean vendor $(BINARY)

.PHONY: clean
clean:
	rm -f $(BINARY)

.PHONY: lint
lint: vendor
	@ echo "\033[36mLinting code\033[0m"
	golangci-lint run --disable-all \
                --exclude-use-default=false \
                --enable=govet \
                --enable=ineffassign \
                --enable=deadcode \
                --enable=golint \
                --enable=goconst \
                --enable=gofmt \
                --enable=goimports \
                --skip-dirs=pkg/client/ \
                --deadline=120s \
                --tests ./...
	@ echo

vendor:
	@ echo "\033[36mPulling dependencies\033[0m"
	GO111MODULE=on GOPROXY=https://proxy.golang.org go mod download
	@ echo

.PHONY: test
test: vendor
	@ echo "\033[36mRunning test suite in Ginkgo\033[0m"
	GO111MODULE=on ginkgo -v -race -randomizeAllSpecs ./... -- -report-dir=$$ARTIFACTS
	@ echo

# Build manager binary
$(BINARY):
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BINARY) -ldflags="-X main.VERSION=${VERSION}"

# Build the docker image
.PHONY: docker-build
docker-build:
	docker build --build-arg VERSION=${VERSION} -t ${IMG}:${VERSION} .
	@echo "\033[36mBuilt $(IMG):$(VERSION)\033[0m"

TAGS ?= latest
.PHONY: docker-tag
docker-tag:
	@IFS=","; tags=${TAGS}; for tag in $${tags}; do docker tag ${IMG}:${VERSION} ${IMG}:$${tag}; echo "\033[36mTagged $(IMG):$(VERSION) as $${tag}\033[0m"; done

# Push the docker image
PUSH_TAGS ?= ${VERSION},latest
.PHONY: docker-push
docker-push:
	@IFS=","; tags=${PUSH_TAGS}; for tag in $${tags}; do docker push ${IMG}:$${tag}; echo "\033[36mPushed $(IMG):$${tag}\033[0m"; done
