# Deis Workflow v2 - End to End Tests

[![Build Status](https://travis-ci.org/deis/workflow-e2e.svg?branch=master)](https://travis-ci.org/deis/workflow-e2e) [![Go Report Card](http://goreportcard.com/badge/deis/workflow-e2e)](http://goreportcard.com/report/deis/workflow-e2e)

Deis (pronounced DAY-iss) is an open source PaaS that makes it easy to deploy and manage
applications on your own servers. Deis builds on [Kubernetes](http://kubernetes.io/) to provide
a lightweight, [Heroku-inspired](http://heroku.com) workflow.

## Work in Progress

![Deis Graphic](https://s3-us-west-2.amazonaws.com/get-deis/deis-graphic-small.png)

Deis Workflow v2 is currently in alpha. Your feedback and participation are more than welcome, but be
aware that this project is considered a work in progress.

## Set up a Deis Cluster

First, install [helm](http://helm.sh) and [boot up a kubernetes cluster][install-k8s]. Next, add the
deis repository to your chart list:

```console
$ helm repo add deis https://github.com/deis/charts
```

Then, install Deis!

```console
$ helm install deis/deis
```

## Run the Tests

To run the entire test suite:

```console
$ ginkgo .
```

To run a single test or set of tests, use the `--focus` option:

```console
$ ginkgo --focus=Apps .
```

## License

Copyright 2015 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


[install-k8s]: http://kubernetes.io/gettingstarted/
