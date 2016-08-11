package tests

import (
	"fmt"

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

// TODO (bacongobbler): inspect kubectl for limits being applied to manifest
var _ = Describe("deis limits", func() {

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

			Specify("that user can list that app's limits", func() {
				sess, err := cmd.Start("deis limits:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say(fmt.Sprintf("=== %s Limits", app.Name)))
				Eventually(sess).Should(Say("--- Memory\nUnlimited"))
				Eventually(sess).Should(Say("--- CPU\nUnlimited"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set a memory limit on that application", func() {
				sess, err := cmd.Start("deis limits:set cmd=64M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     64M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check that --memory also works
				sess, err = cmd.Start("deis limits:set --memory cmd=128M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     128M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set a CPU limit on that application", func() {
				sess, err := cmd.Start("deis limits:set --cpu cmd=500m -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- CPU\ncmd     500m"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can unset a memory limit on that application", func() {
				// no memory has been set
				sess, err := cmd.Start("deis limits:unset cmd -a %s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))

				// Check that --memory also works
				sess, err = cmd.Start("deis limits:set --memory cmd=64M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     64M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
				sess, err = cmd.Start("deis limits:unset --memory cmd -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\nUnlimited"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can unset a CPU limit on that application", func() {
				// no cpu has been set
				sess, err := cmd.Start("deis limits:unset --cpu cmd -a %s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

		})

	})

})
