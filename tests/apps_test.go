package tests

import (
	"os"
	"strings"
	"time"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis apps", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Specify("that user can create an app without a git remote", func() {
			app := apps.Create(user, "--no-remote")
			apps.Destroy(user, app)
		})

		Specify("that user can create an app that uses a custom buildpack", func() {
			app := apps.Create(user, "--no-remote", "--buildpack https://weird-buildpacks.io/lisp")
			defer apps.Destroy(user, app)
			sess, err := cmd.Start("deis config:list -a %s", &user, app.Name)
			Eventually(sess).Should(Say("BUILDPACK_URL"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Context("and an app that does not exist", func() {

			bogusAppName := "bogus-app-name"

			Specify("that user cannot get information about that app", func() {
				sess, err := cmd.Start("deis info -a %s", &user, bogusAppName)
				Eventually(sess.Err).Should(Say("Not found."))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user cannot retrieve logs for that app", func() {
				sess, err := cmd.Start("deis logs -a %s", &user, bogusAppName)
				Eventually(sess.Err).Should(Say(`Error: There are currently no log messages. Please check the following things:`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user cannot open that app", func() {
				sess, err := cmd.Start("deis open -a %s", &user, bogusAppName)
				Eventually(sess.Err).Should(Say("404 Not Found"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user cannot run a command in that app's environment", func() {
				sess, err := cmd.Start("deis apps:run -a %s echo Hello, 世界", &user, bogusAppName)
				Eventually(sess).Should(Say("Running 'echo Hello, 世界'..."))
				Eventually(sess.Err).Should(Say("Not found."))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).ShouldNot(Exit(0))
			})

		})

		Context("who owns an existing app", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user cannot create a new app with the same name", func() {
				sess, err := cmd.Start("deis apps:create %s", &user, app.Name)
				Eventually(sess.Err).Should(Say("App with this id already exists."))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).ShouldNot(Exit(0))
			})

			Context("and another user also exists", func() {

				var otherUser model.User

				BeforeEach(func() {
					otherUser = auth.Register()
				})

				AfterEach(func() {
					auth.Cancel(otherUser)
				})

				Specify("that first user can transfer ownership to the other user", func() {
					sess, err := cmd.Start("deis apps:transfer --app=%s %s", &user, app.Name, otherUser.Username)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
					sess, err = cmd.Start("deis info -a %s", &user, app.Name)
					Eventually(sess.Err).Should(Say("You do not have permission to perform this action."))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(1))
					// Transer back or else cleanup will fail.
					sess, err = cmd.Start("deis apps:transfer --app=%s %s", &otherUser, app.Name, user.Username)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

		Context("who has a local git repo containing source code", func() {

			BeforeEach(func() {
				output, err := cmd.Execute(`git clone https://github.com/deis/example-go.git`)
				Expect(err).NotTo(HaveOccurred(), output)
			})

			Specify("that user can create an app with a git remote", func() {
				os.Chdir("example-go")
				app := apps.Create(user)
				apps.Destroy(user, app)
			})

		})

		Context("who owns an existing app that has already been deployed", func() {

			uuidRegExp := `[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}`
			procsRegexp := `(%s-v\d+-[\w-]+) up \(v\d+\)`

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
				builds.Create(user, app)
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can get information about that app", func() {
				sess, err := cmd.Start("deis info -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Application", app.Name))
				Eventually(sess).Should(Say(`uuid:\s*%s`, uuidRegExp))
				Eventually(sess).Should(Say(`url:\s*%s`, strings.Replace(app.URL, "http://", "", 1)))
				Eventually(sess).Should(Say(`owner:\s*%s`, user.Username))
				Eventually(sess).Should(Say(`id:\s*%s`, app.Name))
				Eventually(sess).Should(Say("=== %s Processes", app.Name))
				Eventually(sess).Should(Say(procsRegexp, app.Name))
				Eventually(sess).Should(Say("=== %s Domains", app.Name))
				Eventually(sess).Should(Say("%s", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			// Pending: The logging subsystem is not currently installed as part of the workflow-dev
			// chart.
			XSpecify("that user can retrieve logs for that app", func() {
				sess, err := cmd.Start("deis logs -a %s", &user, app.Name)
				Eventually(sess).Should(SatisfyAll(
					Say("%s\\[deis-controller\\]\\: %s created initial release", app.Name, user.Username),
					Say("%s\\[deis-controller\\]\\: %s deployed", app.Name, user.Username),
					Say("%s\\[deis-controller\\]\\: %s scaled containers", app.Name, user.Username)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can open that app", func() {
				apps.Open(user, app)
			})

			Specify("that user can run a command in that app's environment", func() {
				sess, err := cmd.Start("deis apps:run --app=%s echo Hello, 世界", &user, app.Name)
				Eventually(sess, (1 * time.Minute)).Should(Say("Hello, 世界"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

		})

	})

})
