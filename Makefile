VERSION		 ?= `git describe --tags --always --dirty`

GO           ?= go
GOTEST		 ?= $(GO) test
GOHOSTOS     ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH   ?= $(shell $(GO) env GOHOSTARCH)
GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)
PREFIX       ?= $(shell pwd)
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GO_LICENSE_DETECTOR ?= $(FIRST_GOPATH)/bin/go-licence-detector

PROMU_VERSION ?= 0.7.0
PROMU         := $(FIRST_GOPATH)/bin/promu-$(PROMU_VERSION)
PROMU_URL     := https://github.com/prometheus/promu/releases/download/v$(PROMU_VERSION)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM).tar.gz

DOCKER_IMAGE_NAME := swoga/ufiber-exporter

ifeq ($(GOHOSTARCH),amd64)
        ifeq ($(GOHOSTOS),$(filter $(GOHOSTOS),linux freebsd darwin windows))
                # Only supported on amd64
                test-flags := -race
        endif
endif

ifeq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	GIT_TREE_STATE=clean
else
	GIT_TREE_STATE=dirty
endif

.PHONY: build
build: promu
	$(PROMU) build --prefix $(PREFIX) $(PROMU_BINARIES)

.PHONY: promu
promu: $(PROMU)

$(PROMU):
	$(eval PROMU_TMP := $(shell mktemp -d))
	curl -s -L $(PROMU_URL) | tar -xvzf - -C $(PROMU_TMP)
	mkdir -p $(FIRST_GOPATH)/bin
	cp $(PROMU_TMP)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM)/promu $(FIRST_GOPATH)/bin/promu-$(PROMU_VERSION)
	rm -r $(PROMU_TMP)

.PHONY: build-release
build-release: clean
	$(PROMU) crossbuild
	$(PROMU) crossbuild tarballs

.PHONY: build-docker
build-docker:
	docker build -t $(DOCKER_IMAGE_NAME) .
	docker tag $(DOCKER_IMAGE_NAME):latest $(DOCKER_IMAGE_NAME):$(VERSION)
	docker tag $(DOCKER_IMAGE_NAME):latest quay.io/$(DOCKER_IMAGE_NAME):latest
	docker tag $(DOCKER_IMAGE_NAME):latest quay.io/$(DOCKER_IMAGE_NAME):$(VERSION)
	docker tag $(DOCKER_IMAGE_NAME):latest ghcr.io/$(DOCKER_IMAGE_NAME):latest
	docker tag $(DOCKER_IMAGE_NAME):latest ghcr.io/$(DOCKER_IMAGE_NAME):$(VERSION)

.PHONY: publish-release
publish-release: check
	git push origin $(VERSION)
	$(PROMU) release .tarballs

.PHONY: docker-hub
docker-hub:
	docker login -u $(DOCKER_HUB_USER) -p $(DOCKER_HUB_PASSWORD)
	docker push $(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_IMAGE_NAME):$(VERSION)

.PHONY: docker-quay
docker-quay:
	docker login quay.io -u $(DOCKER_QUAY_USER) -p $(DOCKER_QUAY_PASSWORD)
	docker push quay.io/$(DOCKER_IMAGE_NAME):latest
	docker push quay.io/$(DOCKER_IMAGE_NAME):$(VERSION)

.PHONY: docker-github
docker-github:
	docker login ghcr.io -u $(GITHUB_USER) -p $(GITHUB_TOKEN)
	docker push ghcr.io/$(DOCKER_IMAGE_NAME):latest
	docker push ghcr.io/$(DOCKER_IMAGE_NAME):$(VERSION)

.PHONY: publish-docker
publish-docker: docker-hub docker-quay docker-github

.PHONY: check
check:
ifeq ($(GIT_TREE_STATE),clean)
	$(info ok)
else
	$(error git state is not clean)
endif

.PHONY: clean
clean:
	rm -rf .build .tarballs
	rm -f ufiber-exporter

.PHONY: release
release: check test build-release build-docker publish-release publish-docker

.PHONY: test
test:
	$(GOTEST) $(test-flags) -v ./...

.PHONY: go-mod-list-updates
go-mod-list-updates:
	$(GO) list -u -m -f '{{if and (not .Indirect) .Update}}{{.}}{{end}}' all

.PHONY: go-licence-detector
go-licence-detector: $(GO_LICENSE_DETECTOR)

$(GO_LICENSE_DETECTOR):
	GO111MODULE=off go get go.elastic.co/go-licence-detector

.PHONY: license-notice
license-notice: go-licence-detector
	$(GO) list -m -json all | $(GO_LICENSE_DETECTOR) -includeIndirect -overrides .licenses/overrides.json -rules .licenses/rules.json -noticeTemplate .licenses/NOTICE.tmpl -noticeOut NOTICE

.PHONY: prepare-release
prepare-release: license-notice
	$(GO) mod tidy