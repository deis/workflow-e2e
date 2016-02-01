package tests

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Apps", func() {
	var appName string

	BeforeEach(func() {
		appName = getRandAppName()
	})

	Context("with no app", func() {

		It("can't get app info", func() {
			sess, _ := start("deis info -a %s", appName)
			Eventually(sess).Should(Exit(1))
			Eventually(sess.Err).Should(Say("Not found."))
		})

		It("can't get app logs", func() {
			sess, err := start("deis logs -a %s", appName)
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(1))
			Eventually(sess.Err).Should(Say("Not found."))
		})

		It("can't run a command in the app environment", func() {
			sess, err := start("deis apps:run echo Hello, 世界")
			Expect(err).To(BeNil())
			Eventually(sess).Should(Say("Running 'echo Hello, 世界'..."))
			Eventually(sess.Err).Should(Say("Not found."))
			Eventually(sess).ShouldNot(Exit(0))
		})

	})

	Context("when creating an app", func() {
		var cleanup bool

		BeforeEach(func() {
			cleanup = true
			appName = getRandAppName()
			cmd, err := start("git init")
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(Say("Initialized empty Git repository"))
		})

		AfterEach(func() {
			if cleanup {
				destroyApp(appName)
				cmd, err := start("rm -rf .git")
				Expect(err).NotTo(HaveOccurred())
				Eventually(cmd).Should(Exit(0))
			}
		})

		It("creates an app with a git remote", func() {
			cmd, err := start("deis apps:create %s", appName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(Say("created %s", appName))
			Eventually(cmd).Should(Say(`Git remote deis added`))
			Eventually(cmd).Should(Say(`remote available at `))
		})

		It("creates an app with no git remote", func() {
			cmd, err := start("deis apps:create %s --no-remote", appName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(SatisfyAll(
				Say("created %s", appName),
				Say("remote available at ")))
			Eventually(cmd).ShouldNot(Say("Git remote deis added"))

			cleanup = false
			cmd = destroyApp(appName)
			Eventually(cmd).ShouldNot(Say("Git remote deis removed"))
		})

		It("creates an app with a custom buildpack", func() {
			sess, err := start("deis apps:create %s --buildpack https://example.com", appName)
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(0))
			Eventually(sess).Should(Say("created %s", appName))
			Eventually(sess).Should(Say("Git remote deis added"))
			Eventually(sess).Should(Say("remote available at "))

			sess, err = start("deis config:list -a %s", appName)
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(0))
			Eventually(sess).Should(Say("BUILDPACK_URL"))
		})
	})

	Context("with a deployed app", func() {
		var appName string

		BeforeEach(func() {
			os.Chdir("example-go")
			appName = getRandAppName()
			cmd := createApp(appName)
			Eventually(cmd).Should(SatisfyAll(
				Say("Git remote deis added"),
				Say("remote available at ")))

			fmt.Println(gitSSH)
			cmd, err := start("GIT_SSH=%s git push deis master", gitSSH)
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd.Err, "2m").Should(Say("done, %s:v2 deployed to Deis", appName))
		})

		AfterEach(func() {
			defer os.Chdir("..")
			destroyApp(appName)
		})

		It("can't create an existing app", func() {
			output, err := execute("deis apps:create %s", appName)
			Expect(err).To(HaveOccurred(), output)

			Expect(output).To(ContainSubstring("This field must be unique"))
		})

		It("can get app info", func() {
			sess, err := start("deis info")
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(Say("=== %s Processes", appName))
			Eventually(sess).Should(SatisfyAny(
				Say("web.1 initialized"),
				Say("web.1 up")))
			Eventually(sess).Should(Say("=== %s Domains", appName))
		})

		// V broken
		XIt("can get app logs", func() {
			cmd, err := start("deis logs")
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(SatisfyAll(
				Say("%s\\[deis-controller\\]\\: %s created initial release", appName, testUser),
				Say("%s\\[deis-controller\\]\\: %s deployed", appName, testUser),
				Say("%s\\[deis-controller\\]\\: %s scaled containers", appName, testUser)))
		})

		// TODO: how to test "deis open" which spawns a browser?
		XIt("can open the app's URL", func() {
			sess, err := start("deis open")
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(0))
		})

		// TODO: be more useful
		XIt("can't open a bogus app URL", func() {
			cmd, err := start("deis open -a %s", getRandAppName())
			Expect(err).To(HaveOccurred())
			Eventually(cmd).Should(Say("404 Not found"))
		})

		// V broken
		XIt("can run a command in the app environment", func() {
			cmd, err := start("deis apps:run echo Hello, 世界")
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd, (1 * time.Minute)).Should(SatisfyAll(
				HavePrefix("Running 'echo Hello, 世界'..."),
				HaveSuffix("Hello, 世界\n")))
		})

		// TODO: this requires a second user account
		XIt("can transfer the app to another owner", func() {
		})
	})
})
