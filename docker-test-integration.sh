#!/usr/bin/env bash
#
# This program is suppsoed to be run inside the workflow-e2e docker container
# to download the workflow CLI at runtime and run the integration tests.

BASE_URL="https://storage.googleapis.com/workflow-cli"
URL="$BASE_URL/deis-latest-linux-amd64"

if [[ $CLI_VERSION -ne "latest" ]]; then
	URL="$BASE_URL/$CLI_VERSION/deis-$CLI_VERSION-linux-amd64"
fi

echo "Installing Workflow CLI version $CLI_VERSION"
curl $URL -o /usr/local/bin/deis && chmod +x /usr/local/bin/deis

make test-integration
