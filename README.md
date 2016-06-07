# Deis Workflow End to End Tests v2

[![Build Status](https://travis-ci.org/deis/workflow-e2e.svg?branch=master)](https://travis-ci.org/deis/workflow-e2e)
[![Go Report Card](http://goreportcard.com/badge/deis/workflow-e2e)](http://goreportcard.com/report/deis/workflow-e2e)
[![Docker Repository on Quay](https://quay.io/repository/deisci/deis-e2e/status "Docker Repository on Quay")](https://quay.io/repository/deisci/deis-e2e)

Deis (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes](http://kubernetes.io) cluster, making it easy to deploy and manage applications on your own servers.

For more information about the Deis Workflow, please visit the main project page at https://github.com/deis/workflow.

We welcome your input! If you have feedback, please [submit an issue][issues]. If you'd like to participate in development, please read the "Development" section below and [submit a pull request][prs].

# About

The code in this repository is a set of [Ginkgo](http://onsi.github.io/ginkgo) and [Gomega](http://onsi.github.io/gomega) based integration tests that execute commands against a running Deis cluster using the Deis CLI.

# Development

The Deis project welcomes contributions from all developers. The high level process for development matches many other open source projects. See below for an outline.

* Fork this repository
* Make your changes
* [Submit a pull request][prs] (PR) to this repository with your changes, and unit tests whenever possible.
  * If your PR fixes any [issues][issues], make sure you write Fixes #1234 in your PR description (where #1234 is the number of the issue you're closing)
* The Deis core contributors will review your code. After each of them sign off on your code, they'll label your PR with LGTM1 and LGTM2 (respectively). Once that happens, the contributors will merge it

## Prerequisities

Before you run the tests, you'll need a full Deis cluster up and running in Kubernetes. Follow the instructions [here](https://github.com/deis/charts#installation) to get one running.

## Run the Tests

There are three options for how to execute the tests. These include two options for executing the tests against Deis Workflow installed on a _remote_ Kubernetes cluster, and one option for installing the same tests directly into a Kubernetes cluster and executing them there.

### Remote Execution

Either of two options for remote execution of the test suite require the `DEIS_CONTROLLER_URL` environment variable to be exported. Its value should be the the controller endpoint you would normally use with the `deis register` or `deis login` commands:

```console
$ export DEIS_CONTROLLER_URL=http://deis.your.cluster
```

Tests execute in parallel by default. If you wish to control the number of executors, export a value for the `GINKGO_NODES` environment variable:

```console
$ export GINKGO_NODES=5
```

If this is not set, Ginkgo will automatically choose a number of test nodes (executors) based on the number of CPU cores _on the machine executing the tests_. It is important to note, however, that test execution is constrained more significantly by the resources of the cluster under test than by the resources of the machine executing the tests. The number of test nodes, therefore, should be explicitly set and scaled in proportion to the resources available in the cluster.

For reference, Workflow's own CI pipeline uses the following:

| Test Nodes | Kubernetes Worker Nodes | Worker Node CPU | Worker Node Memory |
|------------|-------------------------|-----------------|--------------------|
| 5          | 3                       | 4 vCPUs         | 15 GB              |

Setting the `GINKGO_NODES` environment variable to a value of `1` will allow serialized execution of all tests in the suite.

#### Native Execution

If you have Go 1.5 or greater already installed and working properly and also have the [Glide](https://github.com/Masterminds/glide) dependency management tool for Go installed, you may clone this repository into your `$GOPATH`:

```console
git clone git@github.com:deis/workflow-e2e.git $GOPATH/src/github.com/deis/workflow-e2e
```

One-time execution of the following will resolve the test suite's own dependencies:

```console
$ make bootstrap
```

To execute the entire test suite:

```console
$ make test-integration
```

To run a single test or set of tests, you'll need the [Ginkgo](https://github.com/onsi/ginkgo) tool installed on your machine:

```console
$ go get github.com/onsi/ginkgo/ginkgo
```

You can then use the `--focus` option to run subsets of the test suite:

```console
$ ginkgo --focus="deis apps" tests
```

#### Containerized Execution

If you do not have Go 1.5 or greater installed locally, but do have a Docker daemon running locally (or are using docker-machine), you can quite easily execute tests against a remote cluster from within a container.

In this case, you may clone this repository into a path of your own choosing (does not need to be on your `$GOPATH`):

```console
git clone git@github.com:deis/workflow-e2e.git /path/of/your/choice
```

Then build the test image and execute the test suite:

```console
$ make docker-build docker-test-integration
```

### Within the Cluster

A third option is to run the test suite from within the very cluster that is under test.

To install and start the tests:

```console
helm install workflow-beta2-e2e
```

To monitor tests as they execute:

```console
$ kubectl --namespace=deis logs -f workflow-beta2-e2e tests
```

## Special Note on Resetting Cluster State

All tests clean up after themselves, however, in the case of test failures or interruptions, automatic cleanup may not always proceed as intended. This may leave projects, users or other state behind, which may impact future executions of the test suite against the same cluster. (Often all tests will fail.) If you see this behavior, run these commands to clean up. (Replace `deis-workflow-qoxhz` with the name of the deis/workflow pod in your cluster.)

```console
$ kubectl exec -it deis-workflow-qoxhz python manage.py shell
Python 2.7.10 (default, Aug 13 2015, 12:27:27)
[GCC 4.9.2] on linux2
>>> from django.contrib.auth import get_user_model
>>> m = get_user_model()
>>> m.objects.exclude(username='AnonymousUser').delete()
>>> m.objects.all()
```

Note that this is an ongoing issue for which we're planning [a more comprehensive fix](https://github.com/deis/workflow-e2e/issues/12).

## License

Copyright 2015, 2016 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


[install-k8s]: http://kubernetes.io/gettingstarted/
[issues]: https://github.com/deis/workflow-e2e/issues
[prs]: https://github.com/deis/workflow-e2e/pulls
