FROM quay.io/deis/go-dev:0.17.0

ENV K8S_VERSION=1.2.6

RUN go get -u -v github.com/onsi/ginkgo/ginkgo \
	&& curl -o /usr/local/bin/kubectl -Os https://storage.googleapis.com/kubernetes-release/release/v$K8S_VERSION/bin/linux/amd64/kubectl \
	&& chmod +x /usr/local/bin/kubectl

COPY . /go/src/github.com/deis/workflow-e2e

WORKDIR /go/src/github.com/deis/workflow-e2e

RUN glide install

CMD ["./docker-test-integration.sh"]
