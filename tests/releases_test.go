package tests

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis releases", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app that has already been deployed", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
				builds.Create(user, app)
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list that app's releases", func() {
				sess, err := cmd.Start("deis releases:list -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Releases", app.Name))
				Eventually(sess).Should(Say(`v1\s+.*\s+%s created initial release`, user.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can get info on one of the app's releases", func() {
				sess, err := cmd.Start("deis releases:info v2 -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Release v2", app.Name))
				Eventually(sess).Should(Say(`config:\s+[\w-]+`))
				Eventually(sess).Should(Say(`owner:\s+%s`, user.Username))
				Eventually(sess).Should(Say(`summary:\s+%s \w+`, user.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and that app has three releases", func() {

				BeforeEach(func() {
					builds.Create(user, app)
				})

				Specify("that user can roll the application back to the second release", func() {
					sess, err := cmd.Start("deis releases:rollback v2 -a %s", &user, app.Name)
					Eventually(sess).Should(Say(`Rolling back to`))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`...done`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis releases:info v2 -a %s", &user, app.Name)
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Release v2", app.Name))
					Eventually(sess).Should(Say(`config:\s+[\w-]+`))
					Eventually(sess).Should(Say(`owner:\s+%s`, user.Username))
					Eventually(sess).Should(Say(`summary:\s+%s \w+`, user.Username))

					// The updated date has to match a string like 2015-12-22T21:20:31UTC:
					Eventually(sess).Should(Say(`updated:\s+[\w\-\:]+UTC`))
					Eventually(sess).Should(Say(`uuid:\s+[0-9a-f\-]+`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

	})

})
