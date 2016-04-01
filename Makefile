export GO15VENDOREXPERIMENT=1

SHORT_NAME := deis-e2e

SRC_PATH := /go/src/github.com/deis/workflow-e2e
DEV_IMG := quay.io/deis/go-dev:0.9.0

RUN_CMD := docker run --rm -e DEIS_ROUTER_SERVICE_HOST=${DEIS_ROUTER_SERVICE_HOST} -e DEIS_ROUTER_SERVICE_PORT=${DEIS_ROUTER_SERVICE_PORT} -v ${CURDIR}:${SRC_PATH} -w ${SRC_PATH} ${DEV_IMG}
DEV_CMD := docker run --rm -e GO15VENDOREXPERIMENT=1 -v ${CURDIR}:${SRC_PATH} -w ${SRC_PATH} ${DEV_IMG}

TEST_OPTS := -slowSpecThreshold=120.00 -noisyPendings=false
PARALLEL_TEST_OPTS := ${TEST_OPTS} -p

MUTABLE_VERSION ?= canary
VERSION ?= git-$(shell git rev-parse --short HEAD)

DEIS_REGISTRY ?= quay.io/
IMAGE_PREFIX ?= deis
IMAGE := ${DEIS_REGISTRY}${IMAGE_PREFIX}/${SHORT_NAME}:${VERSION}
MUTABLE_IMAGE := ${DEIS_REGISTRY}${IMAGE_PREFIX}/${SHORT_NAME}:${MUTABLE_VERSION}

.PHONY: bootstrap
bootstrap:
	${DEV_CMD} glide install

.PHONY: test-integration
test-integration:
	DEFAULT_EVENTUALLY_TIMEOUT="30s" ginkgo ${TEST_OPTS} tests/

.PHONY: test-integration
test-integration-parallel:
	ginkgo ${PARALLEL_TEST_OPTS} tests/

.PHONY: docker-build
docker-build:
	docker build -t ${IMAGE} ${CURDIR}
	docker tag -f ${IMAGE} ${MUTABLE_IMAGE}

.PHONY: docker-push
docker-push: docker-immutable-push docker-mutable-push

.PHONY: docker-immutable-push
docker-immutable-push:
	docker push ${IMAGE}

.PHONY: docker-mutable-push
docker-mutable-push:
	docker push ${MUTABLE_IMAGE}

.PHONY: docker-test-integration
# run tests inside of a container
docker-test-integration:
	docker run -e DEIS_ROUTER_SERVICE_HOST=${DEIS_ROUTER_SERVICE_HOST} \
	  				 -e DEIS_ROUTER_SERVICE_PORT=${DEIS_ROUTER_SERVICE_PORT} \
						 -e TEST_OPTS=${TEST_OPTS}
						 -e DEFAULT_EVENTUALLY_TIMEOUT="30s" ${IMAGE}

.PHONY: docker-test-integration-parallel
# run tests inside of a container
docker-test-integration-parallel:
	docker run -e DEIS_ROUTER_SERVICE_HOST=${DEIS_ROUTER_SERVICE_HOST} \
	  				 -e DEIS_ROUTER_SERVICE_PORT=${DEIS_ROUTER_SERVICE_PORT} \
						 -e TEST_OPTS=${PARALLEL_TEST_OPTS} ${IMAGE}