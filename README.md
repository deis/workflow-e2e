# Deis End to End Tests v2

[![Build Status](https://travis-ci.org/deis/workflow-e2e.svg?branch=master)](https://travis-ci.org/deis/workflow-e2e)
[![Go Report Card](http://goreportcard.com/badge/deis/workflow-e2e)](http://goreportcard.com/report/deis/workflow-e2e)
[![Docker Repository on Quay](https://quay.io/repository/deisci/deis-e2e/status "Docker Repository on Quay")](https://quay.io/repository/deisci/deis-e2e)

Deis (pronounced DAY-iss) Workflow is an open source Platform as a Service (PaaS) that adds a developer-friendly layer to any [Kubernetes](http://kubernetes.io) cluster, making it easy to deploy and manage applications on your own servers.

For more information about the Deis Workflow, please visit the main project page at https://github.com/deis/workflow.

## Beta Status

This Deis component is currently in beta status, and we welcome your input! If you have feedback, please [submit an issue][issues]. If you'd like to participate in development, please read the "Development" section below and [submit a pull request][prs].

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

To run the entire test suite:

```console
$ make test-integration
```

To run a single test or set of tests, you'll need the [ginkgo](https://github.com/onsi/ginkgo) tool installed. You can then use the `--focus` option:

```console
$ ginkgo --focus=Apps .
```

## Special Note on Resetting Cluster State

Periodically, tests may not clean up after themselves and leave projects, users or other state behind, which will cause lots of test failures (often all tests will fail). If you see this behavior, run these commands to clean up (replace `deis-workflow-qoxhz`) with the name of the deis/workflow pod in your cluster):

```console
$ kubectl exec -it deis-workflow-qoxhz python manage.py shell
Python 2.7.10 (default, Aug 13 2015, 12:27:27)
[GCC 4.9.2] on linux2
>>> from django.contrib.auth import get_user_model
>>> m = get_user_model()
>>> m.objects.exclude(username='AnonymousUser').delete()
>>> m.objects.all()                                     
```

Note that this is an ongoing issue for which we're planning a more comprehensive fix in [this issue](https://github.com/deis/workflow-e2e/issues/12)).

## License

Copyright 2015, 2016 Engine Yard, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


[install-k8s]: http://kubernetes.io/gettingstarted/
[issues]: https://github.com/deis/workflow-e2e/issues
[prs]: https://github.com/deis/workflow-e2e/pulls
