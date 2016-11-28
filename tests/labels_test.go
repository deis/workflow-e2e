package tests

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis labels", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list that app's labels", func() {
				sess, err := cmd.Start("deis labels:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Label", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user cannot set an invalid label", func() {
				sess, err := cmd.Start("deis labels:set --app=%s only_key", &user, app.Name)
				Eventually(sess).ShouldNot(Say(`done`))
				Eventually(sess.Err).Should(Say("only_key is invalid"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user cannot unset an non-exist label", func() {
				sess, err := cmd.Start("deis labels:unset --app=%s not_exist", &user, app.Name)
				Eventually(sess).ShouldNot(Say(`done`))
				Eventually(sess.Err).Should(Say("not_exist does not exist"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user can set a valid tag", func() {
				sess, err := cmd.Start("deis labels:set --app=%s team=bi service=frontend", &user, app.Name)
				Eventually(sess).Should(Say(`Applying labels on %s...`, app.Name))
				Eventually(sess).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis labels:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Label", app.Name))
				Eventually(sess).Should(Say("service:         frontend\nteam:            bi"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can unset that label from that app", func() {
				sess, err := cmd.Start("deis labels:set --app=%s zoo=animal", &user, app.Name)
				Eventually(sess).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis labels:unset --app=%s zoo", &user, app.Name)
				Eventually(sess).Should(Say(`Removing labels on %s...`, app.Name))
				Eventually(sess).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis labels:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Label", app.Name))
				Eventually(sess).ShouldNot(Say("zoo", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and labels has already been added to the app", func() {

				Specify("that user can add more labels to the apps", func() {
					sess, err := cmd.Start("deis labels:set --app=%s team=frontend", &user, app.Name)
					Eventually(sess).Should(Say("done"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis labels:set --app=%s zoo=animal", &user, app.Name)
					Eventually(sess).Should(Say("done"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis labels:list --app=%s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Label", app.Name))
					Eventually(sess).Should(Say("zoo:             animal"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

	})

	DescribeTable("any user can get command-line help for labels", func(command string, expected string) {
		sess, err := cmd.Start(command, nil)
		Eventually(sess).Should(Say(expected))
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Exit(0))
	},
		Entry("helps on \"help labels\"",
			"deis help labels", "Valid commands for labels:"),
		Entry("helps on \"labels -h\"",
			"deis labels -h", "Valid commands for labels:"),
		Entry("helps on \"labels --help\"",
			"deis labels --help", "Valid commands for labels:"),
		Entry("helps on \"help labels:list\"",
			"deis help labels:list", "Prints a list of labels of the application."),
		Entry("helps on \"labels:list -h\"",
			"deis labels:list -h", "Prints a list of labels of the application."),
		Entry("helps on \"labels:list --help\"",
			"deis labels:list --help", "Prints a list of labels of the application."),
		Entry("helps on \"help labels:set\"",
			"deis help labels:set", "Sets labels for an application."),
		Entry("helps on \"labels:set -h\"",
			"deis labels:set -h", "Sets labels for an application."),
		Entry("helps on \"labels:set --help\"",
			"deis labels:set --help", "Sets labels for an application."),
		Entry("helps on \"help labels:unset\"",
			"deis help labels:unset", "Unsets labels for an application."),
		Entry("helps on \"labels:unset -h\"",
			"deis labels:unset -h", "Unsets labels for an application."),
		Entry("helps on \"labels:unset --help\"",
			"deis labels:unset --help", "Unsets labels for an application."),
	)

})
