job('workflow-e2e-pr') {
  description """<ol>
  <li>Watches (Docker-based) repo for pull requests</li>
  <li>Runs _scripts/deploy.sh to build and push Docker image</li>
  <li>Kicks off downstream e2e job to vet changes</li>
</ol>"""
  scm {
    git {
      remote {
        github('deis/workflow-e2e')
        refspec('+refs/pull/*:refs/remotes/origin/pr/*')
      }
      branch('${sha1}')
    }
  }

  publishers {
    slackNotifications {
      projectChannel('#deis-testing')
      notifyAborted()
      notifyFailure()
     }
   }

  parameters {
    stringParam('DOCKER_USERNAME', 'deisbot', 'Docker Hub account name')
    stringParam('DOCKER_EMAIL', 'dummy-address@deis.com', 'Docker Hub account name')
    stringParam('QUAY_USERNAME', 'deisci+jenkins', 'Quay account name')
    stringParam('QUAY_EMAIL', 'deisci+jenkins@deis.com', 'Quay email address')
    stringParam('sha1', 'master', 'Specific Git SHA to test')
  }

  triggers {
    pullRequest {
      admin('deis-admin')
      cron('H/5 * * * *')
      useGitHubHooks()
      // Danger? the following will build all pull requests automatically without asking
      permitAll()
    }
  }

  wrappers {
    timestamps()
    colorizeOutput 'xterm'

    credentialsBinding {
      string("DOCKER_PASSWORD", "0d1f268f-407d-4cd9-a3c2-0f9671df0104")
      string("QUAY_PASSWORD", "c67dc0a1-c8c4-4568-a73d-53ad8530ceeb")
    }
  }

  steps {
    shell '''
      #!/usr/bin/env bash

      set -eo pipefail

      make bootstrap

      export IMAGE_PREFIX=deisci
      docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
      DEIS_REGISTRY='' make docker-build docker-immutable-push
      docker login -e="$QUAY_EMAIL" -u="$QUAY_USERNAME" -p="$QUAY_PASSWORD" quay.io
      DEIS_REGISTRY=quay.io/ make docker-build docker-immutable-push
    '''.stripIndent().trim()

    downstreamParameterized {
      trigger('deis-v2-e2e-pr') {
        parameters {
          predefinedProp('WORKFLOW_E2E_SHA', '${GIT_COMMIT}')
        }
      }
    }
  }
}
