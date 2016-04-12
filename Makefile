export GO15VENDOREXPERIMENT=1

SHORT_NAME := workflow-e2e

SRC_PATH := /go/src/github.com/deis/workflow-e2e

MUTABLE_VERSION ?= canary
VERSION ?= git-$(shell git rev-parse --short HEAD)

ifdef GINKGO_NODES
export GINKO_NODES_ARG=-nodes=${GINKGO_NODES}
else
export GINKO_NODES_ARG=-p
endif

TEST_OPTS := -slowSpecThreshold=120.00 -noisyPendings=false ${GINKO_NODES_ARG}

DEIS_REGISTRY ?= quay.io/
IMAGE_PREFIX ?= deis
IMAGE := ${DEIS_REGISTRY}${IMAGE_PREFIX}/${SHORT_NAME}:${VERSION}
MUTABLE_IMAGE := ${DEIS_REGISTRY}${IMAGE_PREFIX}/${SHORT_NAME}:${MUTABLE_VERSION}

ifndef DEIS_CONTROLLER_URL
ifdef DEIS_ROUTER_SERVICE_HOST
export DEIS_CONTROLLER_URL=http://deis.${DEIS_ROUTER_SERVICE_HOST}.nip.io
endif
endif

DEV_IMG := quay.io/deis/go-dev:0.10.0
DEV_CMD_ARGS := --rm -v ${CURDIR}:${SRC_PATH} -w ${SRC_PATH} ${DEV_IMG}
DEV_CMD := docker run ${DEV_CMD_ARGS}
DEV_CMD_INT := docker run -it ${DEV_CMD_ARGS}
RUN_CMD := docker run --rm -e GINKGO_NODES=${GINKGO_NODES} \
													 -e DEIS_CONTROLLER_URL=${DEIS_CONTROLLER_URL} \
													 -e DEFAULT_EVENTUALLY_TIMEOUT=${DEFAULT_EVENTUALLY_TIMEOUT} \
													 -e MAX_EVENTUALLY_TIMEOUT=${MAX_EVENTUALLY_TIMEOUT} \
													 -e JUNIT=${JUNIT} \
													 -e DEBUG=${DEBUG} \
													 -v ${HOME}/.kube:/root/.kube \
													 -w ${SRC_PATH} ${IMAGE}

check-controller-url:
	@if [ -z "$$DEIS_CONTROLLER_URL" ]; then \
		echo "DEIS_CONTROLLER_URL is not exported. You must export this variable to proceed."; \
		echo "Its value should match the Deis Controller URL you would ordinarily use with"; \
		echo "the \`deis register\` or \`deis login\` commands."; \
	exit 2; \
	fi

dev-env:
	${DEV_CMD_INT} bash

bootstrap:
	glide install

docker-bootstrap:
	${DEV_CMD} make bootstrap

test-integration: check-controller-url
	ginkgo ${TEST_OPTS} tests/

docker-build:
	docker build -t ${IMAGE} ${CURDIR}
	docker tag -f ${IMAGE} ${MUTABLE_IMAGE}

docker-push: docker-immutable-push docker-mutable-push

docker-immutable-push:
	docker push ${IMAGE}

docker-mutable-push:
	docker push ${MUTABLE_IMAGE}

# run tests in parallel inside of a container
docker-test-integration:
	${RUN_CMD} make test-integration

.PHONY: check-controller-url \
				dev-env \
				bootstrap \
				docker-bootstrap \
				test-integration \
				docker-build \
				docker-push \
				docker-immutable-push \
				docker-mutable-push \
				docker-test-integration
