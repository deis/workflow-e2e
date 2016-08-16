#!/usr/bin/env bash
#
# This program is suppsoed to be run inside the workflow-e2e docker container
# to download the workflow CLI at runtime and run the integration tests.
set -eo pipefail

BASE_URL="https://storage.googleapis.com/workflow-cli"
URL="${BASE_URL}/deis-latest-linux-amd64"

if [[ "${CLI_VERSION}" != "latest" ]]; then
	URL="${BASE_URL}/${CLI_VERSION}/deis-${CLI_VERSION}-linux-amd64"
fi

# Download CLI, retry up to 5 times with 10 second delay between each
echo "Installing Workflow CLI version '${CLI_VERSION}' via url '${URL}'"
curl --silent --show-error -I "${URL}"
curl --silent --show-error --retry 5 --retry-delay 10 -o /usr/local/bin/deis "${URL}"
chmod +x /usr/local/bin/deis

echo "Workflow CLI Version '$(deis --version)' installed."

if [ "$TEST" == "bps" ]; then
	make test-buildpacks
	make test-dockerfiles
else
	make test-integration
fi
