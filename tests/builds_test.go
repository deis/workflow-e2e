package tests

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis builds", func() {

	Context("with an existing user", func() {

		uuidRegExp := `[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}`

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("and an app that does not exist", func() {

			bogusAppName := "bogus-app-name"

			Specify("that user cannot create a build for that app", func() {
				sess, err := cmd.Start("deis builds:create -a %s %s", &user, bogusAppName, builds.ExampleImage)
				Eventually(sess.Err).Should(Say(util.PrependError(apps.ErrNoAppMatch)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

		})

		Context("who owns an existing app that has not been deployed", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list that app's builds", func() {
				sess, err := cmd.Start("deis builds:list -a %s", &user, app.Name)
				Eventually(sess).ShouldNot(Say(uuidRegExp))
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
			})

			Specify("that user can create a new build of that app from an existing image", func() {
				builds.Create(user, app)
			})

			Specify("that user can create a new build of that app from an existing image using `deis pull`", func() {
				builds.Pull(user, app)
			})

			Specify("that user can't create a new build of that app from a nonexistent image using `deis pull`", func() {
				builds.Create(user, app)
				// Docker Hub gives a "not found" 400 error
				nonexistentImage := "deis/nonexistent:dummy"
				sess, err := cmd.Start("deis pull --app=%s %s", &user, app.Name, nonexistentImage)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating build..."))
				Eventually(sess.Err).Should(Say(`image .* not found`))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(1))
				// quay.io gives a "permission denied" 400 error
				nonexistentImage = "quay.io/deis/nonexistent:dummy"
				sess, err = cmd.Start("deis pull --app=%s %s", &user, app.Name, nonexistentImage)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating build..."))
				Eventually(sess.Err).Should(Say("Permission Denied attempting to pull image"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(1))

				// test that the old rc is not deleted after a failed build
				procsListing := listProcs(user, app, "").Out.Contents()
				procs := scrapeProcs(app.Name, procsListing)
				Expect(len(procs)).To(Equal(1))
			})

			Specify("that user can create multiple builds of that app with DEPLOY_BATCHES set to 5", func() {
				builds.Pull(user, app)

				// scale to 11
				sess, err := cmd.Start("deis ps:scale cmd=11 --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// configure 5 pods being rolled at once
				sess, err = cmd.Start("deis config:set -a %s DEPLOY_BATCHES=5", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`DEPLOY_BATCHES\s+5`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

			})

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

			Specify("that user can list that app's builds", func() {
				sess, err := cmd.Start("deis builds:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say(uuidRegExp))
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
			})

			Specify("that user can create a new build of that app from an existing image", func() {
				builds.Create(user, app)
			})

			Specify("that user can create a new build of that app from an existing image using `deis pull`", func() {
				builds.Pull(user, app)
			})

		})

	})

})
