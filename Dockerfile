FROM ubuntu-debootstrap:14.04

ENV DOCKERIMAGE=1

RUN apt-get update -y && apt-get install -y curl openssh-client git
RUN curl -sSL http://deis.io/deis-cli/install-v2-alpha.sh | bash && mv ./deis /bin/deis
COPY tests/tests.test .
RUN mv tests.test /bin
RUN mkdir /files
COPY tests/files /files
CMD /bin/tests.test
