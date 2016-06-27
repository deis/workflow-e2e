package tests

import (
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

			Specify("that user can list healthchecks on that app", func() {
				sess, err := cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an exec liveness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set -a %s liveness exec -- /bin/true", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Applying livenessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe: Command=[bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe: Command=[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			// disable httpGet liveness checks as it won't pass for every app
			// TODO(bacongobbler): somehow determine *what* port we need to set the liveness check for this app
			XSpecify("that user can set a httpGet liveness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set liveness httpGet -a %s 80", &user, app.Name)
				Eventually(sess).Should(Say("Applying livenessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`HTTP Get Probe: Path="/" Port=80 HTTPHeaders=[]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`HTTP Get Probe: Path="/" Port=80 HTTPHeaders=[]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set a tcpSocket liveness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set liveness tcpSocket -a %s 80", &user, app.Name)
				Eventually(sess).Should(Say("Applying livenessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`TCP Socket Probe: Port=80`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`TCP Socket Probe: Port=80`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an exec readiness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set readiness exec -a %s -- /bin/true", &user, app.Name)
				Eventually(sess).Should(Say("Applying readinessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe: Command=[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe: Command=[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set a httpGet readiness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set readiness httpGet -a %s 80", &user, app.Name)
				Eventually(sess).Should(Say("Applying readinessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`HTTP Get Probe: Path="/" Port=80 HTTPHeaders=[]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`HTTP Get Probe: Path="/" Port=80 HTTPHeaders=[]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set a tcpSocket readiness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set readiness tcpSocket -a %s 80", &user, app.Name)
				Eventually(sess).Should(Say("Applying readinessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`TCP Socket Probe: Port=80`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`TCP Socket Probe: Port=80`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and already has a healthcheck set", func() {
				BeforeEach(func() {
					sess, err := cmd.Start(`deis healthchecks:set readiness exec -a %s -- /bin/true`, &user, app.Name)
					Eventually(sess).Should(Say("Applying readinessProbe healthcheck..."))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
					Eventually(sess).Should(Say(`Exec Probe: Command=[/bin/true]`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can unset that healthcheck", func() {
					sess, err := cmd.Start("deis healthcheck:unset -a %s readiness", &user, app.Name)
					Eventually(sess).Should(Say("Removing healthchecks..."))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
					Eventually(sess).ShouldNot(Say(`Exec Probe: Command=[/bin/true]`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
					Eventually(sess).ShouldNot(Say(`Exec Probe: Command=[/bin/true]`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})
			})
		})
	})

	DescribeTable("any user can get command-line help for healthchecks", func(command string, expected string) {
		sess, err := cmd.Start(command, nil)
		Eventually(sess).Should(Say(expected))
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Exit(0))
		// TODO: test that help output was more than five lines long
	},
		Entry("helps on \"help healthchecks\"",
			"deis help healthchecks", "Valid commands for healthchecks:"),
		Entry("helps on \"healthchecks -h\"",
			"deis healthchecks -h", "Valid commands for healthchecks:"),
		Entry("helps on \"healthchecks --help\"",
			"deis healthchecks --help", "Valid commands for healthchecks:"),
		Entry("helps on \"help healthchecks:list\"",
			"deis help healthchecks:list", "Lists healthchecks for an application."),
		Entry("helps on \"healthchecks:list -h\"",
			"deis healthchecks:list -h", "Lists healthchecks for an application."),
		Entry("helps on \"healthchecks:list --help\"",
			"deis healthchecks:list --help", "Lists healthchecks for an application."),
		Entry("helps on \"help healthchecks:set\"",
			"deis help healthchecks:set", "Sets healthchecks for an application."),
		Entry("helps on \"healthchecks:set -h\"",
			"deis healthchecks:set -h", "Sets healthchecks for an application."),
		Entry("helps on \"healthchecks:set --help\"",
			"deis healthchecks:set --help", "Sets healthchecks for an application."),
		Entry("helps on \"help healthchecks:unset\"",
			"deis help healthchecks:unset", "Unsets healthchecks for an application."),
		Entry("helps on \"healthchecks:unset -h\"",
			"deis healthchecks:unset -h", "Unsets healthchecks for an application."),
		Entry("helps on \"healthchecks:unset --help\"",
			"deis healthchecks:unset --help", "Unsets healthchecks for an application."),
	)
})
