FROM quay.io/deis/go-dev:0.9.0

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl wget openssh-client git g++ gcc libc6-dev make bzr git mercurial openssh-client subversion procps && \
    rm -rf /var/lib/apt/lists/*

#Install deis cli
RUN wget https://dl.bintray.com/deis/deisci/deis-6c176d2-linux-amd64 && \
    mv ./deis-6c176d2-linux-amd64 /bin/deis && \
    chmod +x /bin/deis

# Install the kubectl cli
RUN curl -O https://storage.googleapis.com/kubernetes-release/release/v1.1.8/bin/linux/amd64/kubectl && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/kubectl

COPY . /go/src/github.com/deis/workflow-e2e
WORKDIR /go/src/github.com/deis/workflow-e2e
