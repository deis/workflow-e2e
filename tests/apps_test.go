package tests

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/deis/workflow-e2e/shims"
)

var _ = Describe("Apps", func() {
	var appName string
	var appURL string

	BeforeEach(func() {
		appName = getRandAppName()
		appURL = strings.Replace(url, "deis", appName, 1)
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
			Eventually(sess.Err).Should(Say(`Error: There are currently no log messages. Please check the following things:`))
		})

		It("can't run a command in the app environment", func() {
			sess, err := start("deis apps:run echo Hello, 世界")
			Expect(err).To(BeNil())
			Eventually(sess).Should(Say("Running 'echo Hello, 世界'..."))
			Eventually(sess.Err).Should(Say("Not found."))
			Eventually(sess).ShouldNot(Exit(0))
		})

		It("can't open a bogus app URL", func() {
			sess, err := start("deis open -a %s", getRandAppName())
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(1))
			Eventually(sess.Err).Should(Say("404 Not Found"))
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
			Eventually(cmd).Should(Exit(0))
			cmd, err := start("GIT_SSH=%s git push deis master", gitSSH)
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd.Err, "2m").Should(Say("Done, %s:v2 deployed to Deis", appName))
			Eventually(cmd).Should(Exit(0))
		})

		AfterEach(func() {
			defer os.Chdir("..")
			destroyApp(appName)
		})

		It("can't create an existing app", func() {
			sess, err := start("deis apps:create %s", appName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("App with this id already exists."))
			Eventually(sess).ShouldNot(Exit(0))
		})

		It("can get app info", func() {
			sess, err := start("deis info -a %s", appName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Processes", appName))
			Eventually(sess).Should(SatisfyAny(
				Say("web.1 initialized"),
				Say("web.1 up")))
			Eventually(sess).Should(Say("=== %s Domains", appName))
			Eventually(sess).Should(Exit(0))
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

		XIt("can open the app's URL", func() {
			// the underlying open utility 'deis open' looks for
			toShim := "open" //darwin
			if runtime.GOOS == "linux" {
				toShim = "xdg-open"
			}
			myShim, err := shims.CreateSystemShim(toShim)
			if err != nil {
				panic(err)
			}
			defer shims.RemoveShim(myShim)

			// create custom env with custom/prefixed PATH value
			env := shims.PrependPath(os.Environ(), os.TempDir())

			// invoke functionality under test
			sess, err := startCmd(Cmd{Env: env, CommandLineString: "deis open"})
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(0))

			// check shim output
			output, err := ioutil.ReadFile(myShim.OutFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(appURL))
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
