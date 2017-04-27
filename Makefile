.PHONY: attack clean darwin dockerise linux lint report

DURATION ?= 10m

FRONTEND := frontend
BACKEND := backend

# This repo's root import path (under GOPATH).
PKG := github.com/jabley/monitoring-spike

# Where to push the docker image.
REGISTRY ?= jabley

# Which architecture to build - see $(ALL_ARCH) for options.
ARCH ?= amd64

# This version-strategy uses git tags to set the version string
# VERSION := $(shell git describe --tags --always --dirty)

# This version-strategy uses semver
VERSION := 1.0.0

###
### These variables should not need tweaking.
###

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

ALL_ARCH := amd64 arm arm64 ppc64le

# Set default base image dynamically for each arch
ifeq ($(ARCH),amd64)
    BASEIMAGE?=alpine
endif
ifeq ($(ARCH),arm)
    BASEIMAGE?=armel/busybox
endif
ifeq ($(ARCH),arm64)
    BASEIMAGE?=aarch64/busybox
endif
ifeq ($(ARCH),ppc64le)
    BASEIMAGE?=ppc64le/busybox
endif

# IMAGE := $(REGISTRY)/$(BIN)-$(ARCH)
BACKEND_IMAGE  := $(REGISTRY)/$(BACKEND)-$(ARCH)
FRONTEND_IMAGE := $(REGISTRY)/$(FRONTEND)-$(ARCH)

BUILD_IMAGE ?= golang:1.8-alpine

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: build

build-%:
	@$(MAKE) --no-print-directory ARCH=$* build

container-%:
	@$(MAKE) --no-print-directory ARCH=$* container

push-%:
	@$(MAKE) --no-print-directory ARCH=$* push

all-build: $(addprefix build-, $(ALL_ARCH))

all-container: $(addprefix container-, $(ALL_ARCH))

all-push: $(addprefix push-, $(ALL_ARCH))

build: bin/$(ARCH)/$(BACKEND) bin/$(ARCH)/$(FRONTEND)

bin/$(ARCH)/%: build-dirs
	@echo "building: $@"
	@docker run                                                            \
	    -ti                                                                \
	    -u $$(id -u):$$(id -g)                                             \
	    -v $$(pwd)/.go:/go                                                 \
	    -v $$(pwd):/go/src/$(PKG)                                          \
	    -v $$(pwd)/bin/$(ARCH):/go/bin                                     \
	    -v $$(pwd)/bin/$(ARCH):/go/bin/linux_$(ARCH)                       \
	    -v $$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static  \
	    -w /go/src/$(PKG)                                                  \
	    $(BUILD_IMAGE)                                                     \
	    /bin/sh -c "                                                       \
	        ARCH=$(ARCH)                                                   \
	        VERSION=$(VERSION)                                             \
	        PKG=$(PKG)                                                     \
					BIN=$*																										     \
	        ./build/build.sh                                               \
	    "

DOTFILE_BACKEND_IMAGE = $(subst :,_,$(subst /,_,$(BACKEND_IMAGE))-$(VERSION))
DOTFILE_FRONTEND_IMAGE = $(subst :,_,$(subst /,_,$(FRONTEND_IMAGE))-$(VERSION))

container: .container-$(DOTFILE_BACKEND_IMAGE) .container-$(DOTFILE_FRONTEND_IMAGE) container-name

.container-$(DOTFILE_BACKEND_IMAGE): bin/$(ARCH)/$(BACKEND) Dockerfile.in
	@sed \
	    -e 's|ARG_BIN|$(BACKEND)|g' \
	    -e 's|ARG_ARCH|$(ARCH)|g' \
	    -e 's|ARG_FROM|$(BASEIMAGE)|g' \
	    Dockerfile.in > .dockerfile-$(BACKEND)-$(ARCH)
	@docker build -t $(BACKEND_IMAGE):$(VERSION) -f .dockerfile-$(BACKEND)-$(ARCH) .
	@docker images -q $(BACKEND_IMAGE):$(VERSION) > $@

.container-$(DOTFILE_FRONTEND_IMAGE): bin/$(ARCH)/$(FRONTEND) Dockerfile.in
	@sed \
	    -e 's|ARG_BIN|$(FRONTEND)|g' \
	    -e 's|ARG_ARCH|$(ARCH)|g' \
	    -e 's|ARG_FROM|$(BASEIMAGE)|g' \
	    Dockerfile.in > .dockerfile-$(FRONTEND)-$(ARCH)
	@docker build -t $(FRONTEND_IMAGE):$(VERSION) -f .dockerfile-$(FRONTEND)-$(ARCH) .
	@docker images -q $(FRONTEND_IMAGE):$(VERSION) > $@

container-name:
	@echo "container: $(BACKEND_IMAGE):$(VERSION)"
	@echo "container: $(FRONTEND_IMAGE):$(VERSION)"

push: .push-$(DOTFILE_BACKEND_IMAGE) .push-$(DOTFILE_FRONTEND_IMAGE) .push-$(DOTFILE_BACKEND_IMAGE) push-name
.push-$(DOTFILE_BACKEND_IMAGE): .container-$(DOTFILE_BACKEND_IMAGE)
ifeq ($(findstring gcr.io,$(REGISTRY)),gcr.io)
	@gcloud docker -- push $(BACKEND_IMAGE):$(VERSION)
else
	@docker push $(BACKEND_IMAGE):$(VERSION)
endif
	@docker images -q $(BACKEND_IMAGE):$(VERSION) > $@

.push-$(DOTFILE_FRONTEND_IMAGE): .container-$(DOTFILE_FRONTEND_IMAGE)
	ifeq ($(findstring gcr.io,$(REGISTRY)),gcr.io)
		@gcloud docker -- push $(FRONTEND_IMAGE):$(VERSION)
	else
		@docker push $(FRONTEND_IMAGE):$(VERSION)
	endif
		@docker images -q $(FRONTEND_IMAGE):$(VERSION) > $@

push-name:
	@echo "pushed: $(BACKEND_IMAGE):$(VERSION)"
	@echo "pushed: $(FRONTEND_IMAGE):$(VERSION)"

version:
	@echo $(VERSION)

test: build-dirs
	@docker run                                                            \
	    -ti                                                                \
	    -u $$(id -u):$$(id -g)                                             \
	    -v $$(pwd)/.go:/go                                                 \
	    -v $$(pwd):/go/src/$(PKG)                                          \
	    -v $$(pwd)/bin/$(ARCH):/go/bin                                     \
	    -v $$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static  \
	    -w /go/src/$(PKG)                                                  \
	    $(BUILD_IMAGE)                                                     \
	    /bin/sh -c "                                                       \
	        ./build/test.sh $(SRC_DIRS)                                    \
	    "

build-dirs:
	@mkdir -p bin/$(ARCH)
	@mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(ARCH)

clean: container-clean bin-clean

container-clean:
	rm -rf .container-* .dockerfile-* .push-*

bin-clean:
	rm -rf .go bin

lint:
	golint ./...
	go vet ./...

run: container
	@docker-compose up -d

stop:
	@docker-compose down
	
attack: run
	vegeta attack -targets=targets.txt -duration=$(DURATION) -rate=50 > results.bin

report: attack
	cat results.bin | vegeta report
