package tests

import (
	"io/ioutil"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis config", func() {

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

			Specify("that user can list environment variables on that app", func() {
				sess, err := cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set environment variables on that app", func() {
				sess, err := cmd.Start("deis config:set -a %s POWERED_BY=midi-chlorians", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`POWERED_BY\s+midi-chlorians`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`POWERED_BY\s+midi-chlorians`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run env -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("POWERED_BY=midi-chlorians"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set multiple environment variables at once on that app", func() {
				sess, err := cmd.Start("deis config:set FOO=null BAR=nil -a %s", &user, app.Name)
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				output := string(sess.Out.Contents())
				Expect(output).To(MatchRegexp(`FOO\s+null`))
				Expect(output).To(MatchRegexp(`BAR\s+nil`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				output = string(sess.Out.Contents())
				Expect(output).To(MatchRegexp(`FOO\s+null`))
				Expect(output).To(MatchRegexp(`BAR\s+nil`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run env -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("FOO=null"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("BAR=nil"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an environment variable containing spaces on that app", func() {
				sess, err := cmd.Start(`deis config:set -a %s POWERED_BY=the\ Deis\ team`, &user, app.Name)
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`POWERED_BY\s+the Deis team`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`POWERED_BY\s+the Deis team`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run -a %s env", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("POWERED_BY=the Deis team"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			It("that user can set a multi-line environment variable on that app", func() {
				value := `This is
a
multiline string.`

				sess, err := cmd.Start(`deis config:set -a %s FOO='%s'`, &user, app.Name, value)
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`FOO\s+%s`, value))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`FOO\s+%s`, value))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run -a %s env", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("FOO=%s", value))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an environment variable with non-ASCII and multibyte chars on that app", func() {
				sess, err := cmd.Start("deis config:set FOO=讲台 BAR=Þorbjörnsson BAZ=ноль -a %s", &user, app.Name)
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				output := string(sess.Out.Contents())
				Expect(output).To(MatchRegexp(`FOO\s+讲台`))
				Expect(output).To(MatchRegexp(`BAR\s+Þorbjörnsson`))
				Expect(output).To(MatchRegexp(`BAZ\s+ноль`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				output = string(sess.Out.Contents())
				Expect(output).To(MatchRegexp(`FOO\s+讲台`))
				Expect(output).To(MatchRegexp(`BAR\s+Þorbjörnsson`))
				Expect(output).To(MatchRegexp(`BAZ\s+ноль`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run -a %s env", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
				output = string(sess.Out.Contents())
				Expect(output).To(ContainSubstring("FOO=讲台"))
				Expect(output).To(ContainSubstring("BAR=Þorbjörnsson"))
				Expect(output).To(ContainSubstring("BAZ=ноль"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and has already has an environment variable set", func() {

				BeforeEach(func() {
					sess, err := cmd.Start(`deis config:set -a %s FOO=xyzzy`, &user, app.Name)
					Eventually(sess).Should(Say("Creating config"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
					Eventually(sess).Should(Say(`FOO\s+xyzzy`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can unset that environment variable", func() {
					sess, err := cmd.Start("deis config:unset -a %s FOO", &user, app.Name)
					Eventually(sess).Should(Say("Removing config"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
					Eventually(sess).ShouldNot(Say(`FOO\s+xyzzy`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Config", app.Name))
					Eventually(sess).ShouldNot(Say(`FOO\s+xyzzy`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis run -a %s env", &user, app.Name)
					Eventually(sess, settings.MaxEventuallyTimeout).ShouldNot(Say("FOO=xyzzy"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can pull the configuration to an .env file", func() {
					sess, err := cmd.Start("deis config:pull -a %s", &user, app.Name)
					// TODO: ginkgo seems to redirect deis' file output here, so just examine
					// the output stream rather than reading in the .env file. Bug?
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("FOO=xyzzy"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

			Specify("that user can push configuration from an .env file", func() {
				contents := []byte(`BIP=baz
FOO=bar`)
				err := ioutil.WriteFile(".env", contents, 0644)
				Expect(err).NotTo(HaveOccurred())

				sess, err := cmd.Start("deis config:push -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Config should appear in config:list.
				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`BIP\s+baz`))
				Eventually(sess).Should(Say(`FOO\s+bar`))

				// Config should be found within the app env vars (without any line endings).
				sess, err = cmd.Start("deis run -a %s 'printf %%q $BIP'", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
				Eventually(sess).Should(Say("baz"))
			})

			Specify("that user can push configuration from an .env file with CRLF line endings", func() {
				contents := []byte("BIP=baz\r\nFOO=bar\r\nWOO=goo\r\n")
				err := ioutil.WriteFile(".env", contents, 0644)
				Expect(err).NotTo(HaveOccurred())

				sess, err := cmd.Start("deis config:push -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Config should appear in the config:list
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`BIP\s+baz`))
				Eventually(sess).Should(Say(`FOO\s+bar`))
				Eventually(sess).Should(Say(`WOO\s+goo`))
				Eventually(sess).Should(Exit(0))

				// Config should be found within the app env vars (without any line endings).
				sess, err = cmd.Start("deis run -a %s 'printf %%q $WOO'", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
				Eventually(sess).Should(Say("goo"))
			})

		})

	})

	DescribeTable("any user can get command-line help for config", func(command string, expected string) {
		sess, err := cmd.Start(command, nil)
		Eventually(sess).Should(Say(expected))
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Exit(0))
		// TODO: test that help output was more than five lines long
	},
		Entry("helps on \"help config\"",
			"deis help config", "Valid commands for config:"),
		Entry("helps on \"config -h\"",
			"deis config -h", "Valid commands for config:"),
		Entry("helps on \"config --help\"",
			"deis config --help", "Valid commands for config:"),
		Entry("helps on \"help config:list\"",
			"deis help config:list", "Lists environment variables for an application."),
		Entry("helps on \"config:list -h\"",
			"deis config:list -h", "Lists environment variables for an application."),
		Entry("helps on \"config:list --help\"",
			"deis config:list --help", "Lists environment variables for an application."),
		Entry("helps on \"help config:set\"",
			"deis help config:set", "Sets environment variables for an application."),
		Entry("helps on \"config:set -h\"",
			"deis config:set -h", "Sets environment variables for an application."),
		Entry("helps on \"config:set --help\"",
			"deis config:set --help", "Sets environment variables for an application."),
		Entry("helps on \"help config:unset\"",
			"deis help config:unset", "Unsets an environment variable for an application."),
		Entry("helps on \"config:unset -h\"",
			"deis config:unset -h", "Unsets an environment variable for an application."),
		Entry("helps on \"config:unset --help\"",
			"deis config:unset --help", "Unsets an environment variable for an application."),
		Entry("helps on \"help config:pull\"",
			"deis help config:pull", "Extract all environment variables from an application for local use."),
		Entry("helps on \"config:pull -h\"",
			"deis config:pull -h", "Extract all environment variables from an application for local use."),
		Entry("helps on \"config:pull --help\"",
			"deis config:pull --help", "Extract all environment variables from an application for local use."),
		Entry("helps on \"help config:push\"",
			"deis help config:push", "Sets environment variables for an application."),
		Entry("helps on \"config:push -h\"",
			"deis config:push -h", "Sets environment variables for an application."),
		Entry("helps on \"config:push --help\"",
			"deis config:push --help", "Sets environment variables for an application."),
	)

})
