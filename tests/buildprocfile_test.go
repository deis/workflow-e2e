package tests

import (
	"time"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("deis builds procfile", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app that has not been deployed", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})
			// https://github.com/deis/workflow-e2e/issues/299
			// TODO(smothiki): somehow determine why this test is flaky in jenkins after merging deployments
			XSpecify("that user create a new build of that app with a different procfile", func() {
				// Docker Hub gives a "not found" 400 error
				Image := "smothiki/exampleapp:latest"
				procfile := "web: /bin/boot"
				sess, err := cmd.Start("deis pull %s --app=%s ", &user, Image, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating build..."))
				time.Sleep(30000 * time.Millisecond)
				sess, err = cmd.Start("deis builds:create %s -a %s --procfile \"%s\"", &user, Image, app.Name, procfile)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating build..."))
				time.Sleep(20000 * time.Millisecond)
				sess, err = cmd.Start("deis ps:scale web=1 -a %s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				time.Sleep(20000 * time.Millisecond)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", app.Name))
				Eventually(sess).Should(Say("--- %s:", "web"))
				Eventually(sess).Should(Say(`(%s-[\w-]+) up \(v\d+\)`, app.Name))

			})

		})

	})

})
