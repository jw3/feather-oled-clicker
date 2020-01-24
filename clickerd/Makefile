# structure from argoproj

CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE}

#  docker image publishing options
DOCKER_PUSH?=true
IMAGE_NAMESPACE?=jwiii
IMAGE_TAG?=latest
GOARCH?=amd64
CGO_ENABLED?=0
GOOS?=linux

ifeq (${DOCKER_PUSH},true)
ifndef IMAGE_NAMESPACE
$(error IMAGE_NAMESPACE must be set to push images (e.g. IMAGE_NAMESPACE=argoproj))
endif
endif

ifneq (${GIT_TAG},)
IMAGE_TAG=${GIT_TAG}
override LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif

ifdef IMAGE_NAMESPACE
IMAGE_PREFIX=${IMAGE_NAMESPACE}/
endif

ifeq (${GOARCH},arm)
export GOARM=7
endif

.DELETE_ON_ERROR:
all: clickerd cli

all-images: clickerd-image

.PHONY: all clean clickerd cli

clickerd:
	env | grep GO
	go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/clickerd ./cmd/main.go

cli:
	env | grep GO
	go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/clicker ./cli/main.go

clickerd-image:
	docker build -t $(IMAGE_PREFIX)clickerd:$(IMAGE_TAG) -f ./servers/cmd/Dockerfile .
	@if [ "$(DOCKER_PUSH)" = "true" ] ; then  docker push $(IMAGE_PREFIX)clickerd:$(IMAGE_TAG) ; fi

clean:
	@if [ -f ${DIST_DIR}/clicker ] ; then rm -v ${DIST_DIR}/clicker ; fi
	@if [ -f ${DIST_DIR}/clickerd ] ; then rm -v ${DIST_DIR}/clickerd ; fi
