#!/usr/bin/env bash
#
# This program is suppsoed to be run inside the workflow-e2e docker container
# to download the workflow CLI at runtime and run the integration tests.
set -eo pipefail

function debug {
	if [ "${DEBUG_MODE}" == "true" ]; then
		filename="/tmp/deis_debug"
		touch "${filename}"
		echo "Sleeping until ${filename} is deleted"

		while [ -f "${filename}" ]
		do
			sleep 2
		done
	fi
}

trap debug ERR

curl-cli-from-gcs-bucket() {
	local gcs_bucket="${1}"
	local base_url="https://storage.googleapis.com/${gcs_bucket}"
	local url

	case "${CLI_VERSION}" in
		"latest" | "stable")
			url="${base_url}"
			;;
		*)
			url="${base_url}/${CLI_VERSION}"
			;;
	esac
	url="${url}/deis-${CLI_VERSION}-linux-amd64"

	# Download CLI, retry up to 5 times with 10 second delay between each
	echo "Installing Workflow CLI version '${CLI_VERSION}' via url '${url}'"
	curl -f --silent --show-error -I "${url}"
	curl -f --silent --show-error --retry 5 --retry-delay 10 -o /usr/local/bin/deis "${url}"
}

# try multiple buckets for specific CLI_VERSION
curl-cli-from-gcs-bucket "workflow-cli-master" || \
curl-cli-from-gcs-bucket "workflow-cli-pr" || \
curl-cli-from-gcs-bucket "workflow-cli-release"
chmod +x /usr/local/bin/deis

echo "Workflow CLI Version '$(deis --version)' installed."

if [ "$TEST" == "bps" ]; then
	make test-buildpacks
	make test-dockerfiles
else
	make test-integration
fi
