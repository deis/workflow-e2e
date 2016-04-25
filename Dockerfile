FROM golang:1.6-alpine

ENV K8S_VERSION=1.2.2 GLIDE_VERSION=0.10.2 GLIDE_HOME=/root GO15VENDOREXPERIMENT=1 JUNIT=true

RUN apk add --update-cache \
	bash \
	curl \
	git \
	make \
	openssh \
	&& rm -rf /var/cache/apk/* \
	&& go get -u -v \
	github.com/tools/godep \
	github.com/onsi/ginkgo/ginkgo \
	&& curl -L https://dl.bintray.com/deis/deisci/deis-852b0b0-linux-amd64 -o /usr/local/bin/deis \
	&& chmod +x /usr/local/bin/deis \
	&& mkdir -p $GOPATH/src/k8s.io \
	&& curl -L https://github.com/kubernetes/kubernetes/archive/v$K8S_VERSION.tar.gz | tar xvz -C $GOPATH/src/k8s.io \
	&& mv $GOPATH/src/k8s.io/kubernetes-$K8S_VERSION $GOPATH/src/k8s.io/kubernetes \
	&& cd $GOPATH/src/k8s.io/kubernetes \
	&& CGO_ENABLED=0 godep go build -o /usr/local/bin/kubectl cmd/kubectl/kubectl.go \
	&& cd ~ \
	&& rm -rf $GOPATH/src/k8s.io/kubernetes \
	&& curl -L https://github.com/Masterminds/glide/releases/download/$GLIDE_VERSION/glide-$GLIDE_VERSION-linux-amd64.tar.gz | tar xvz -C /tmp \
	&& mv /tmp/linux-amd64/glide /usr/local/bin \ 
	&& rm -rf /tmp/linux-amd64

COPY . /go/src/github.com/deis/workflow-e2e

WORKDIR /go/src/github.com/deis/workflow-e2e

RUN glide install

CMD ["/usr/bin/make", "test-integration"]
