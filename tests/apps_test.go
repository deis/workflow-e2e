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

var uuidRegExp = `[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}`

var _ = Describe("Apps", func() {
	var testApp App

	BeforeEach(func() {
		url, testUser, testPassword, testEmail, keyName = createRandomUser()
		testApp.Name = getRandAppName()
	})

	Context("with no app", func() {

		It("can't get app info", func() {
			sess, _ := start("deis info -a %s", testApp.Name)
			Eventually(sess).Should(Exit(1))
			Eventually(sess.Err).Should(Say("Not found."))
		})

		It("can't get app logs", func() {
			sess, err := start("deis logs -a %s", testApp.Name)
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

		BeforeEach(func() {
			url, testUser, testPassword, testEmail, keyName = createRandomUser()
			testApp.Name = getRandAppName()
			gitInit()
		})

		It("creates an app with a git remote", func() {
			cmd, err := start("deis apps:create %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(Say("created %s", testApp.Name))
			Eventually(cmd).Should(Say(`Git remote deis added`))
			Eventually(cmd).Should(Say(`remote available at `))
		})

		It("creates an app with no git remote", func() {
			cmd, err := start("deis apps:create %s --no-remote", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(SatisfyAll(
				Say("created %s", testApp.Name),
				Say("remote available at ")))
			Eventually(cmd).ShouldNot(Say("Git remote deis added"))
		})

		It("creates an app with a custom buildpack", func() {
			sess, err := start("deis apps:create %s --buildpack https://example.com", testApp.Name)
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(0))
			Eventually(sess).Should(Say("created %s", testApp.Name))
			Eventually(sess).Should(Say("Git remote deis added"))
			Eventually(sess).Should(Say("remote available at "))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(0))
			Eventually(sess).Should(Say("BUILDPACK_URL"))
		})
	})

	Context("with a deployed app", func() {
		var testApp App

		BeforeEach(func() {
			url, testUser, testPassword, testEmail, keyName = createRandomUser()
			os.Chdir("example-go")
			appName := getRandAppName()
			createApp(appName)
			testApp = deployApp(appName)
		})

		AfterEach(func() {
			defer os.Chdir("..")
		})

		It("can't create an existing app", func() {
			sess, err := start("deis apps:create %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("App with this id already exists."))
			Eventually(sess).ShouldNot(Exit(0))
		})

		It("can get app info", func() {
			verifyAppInfo(testApp)
		})

		// V broken
		XIt("can get app logs", func() {
			cmd, err := start("deis logs")
			Expect(err).NotTo(HaveOccurred())
			Eventually(cmd).Should(SatisfyAll(
				Say("%s\\[deis-controller\\]\\: %s created initial release", testApp.Name, testUser),
				Say("%s\\[deis-controller\\]\\: %s deployed", testApp.Name, testUser),
				Say("%s\\[deis-controller\\]\\: %s scaled containers", testApp.Name, testUser)))
		})

		It("can open the app's URL", func() {
			verifyAppOpen(testApp)
		})

		It("can run a command in the app environment", func() {
			sess, err := start("deis apps:run echo Hello, 世界")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, (1 * time.Minute)).Should(Say("Hello, 世界"))
		})

		It("can transfer the app to another owner", func() {
			_, err := start("deis apps:transfer " + testAdminUser)
			Expect(err).NotTo(HaveOccurred())
			sess, _ := start("deis info -a %s", testApp.Name)
			Eventually(sess).Should(Exit(1))
			Eventually(sess.Err).Should(Say("You do not have permission to perform this action."))
		})
	})

	Context("with a custom buildpack deployed app", func() {
		var testApp App

		BeforeEach(func() {
			url, testUser, testPassword, testEmail, keyName = createRandomUser()
			os.Chdir("example-perl")
			appName := getRandAppName()
			createApp(appName, "--buildpack", "https://github.com/miyagawa/heroku-buildpack-perl.git")
			testApp = deployApp(appName)
		})

		It("can get app info", func() {
			verifyAppInfo(testApp)
		})

		It("can open the app's URL", func() {
			verifyAppOpen(testApp)
		})

	})
})

func verifyAppInfo(testApp App) {
	sess, err := start("deis info -a %s", testApp.Name)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("=== %s Application", testApp.Name))
	Eventually(sess).Should(Say(`uuid:\s*%s`, uuidRegExp))
	Eventually(sess).Should(Say(`url:\s*%s`, strings.Replace(testApp.URL, "http://", "", 1)))
	Eventually(sess).Should(Say(`owner:\s*%s`, testUser))
	Eventually(sess).Should(Say(`id:\s*%s`, testApp.Name))

	Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
	Eventually(sess).Should(Say(procsRegexp, testApp.Name))

	Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
	Eventually(sess).Should(Say("%s", testApp.Name))
	Eventually(sess).Should(Exit(0))
}

func verifyAppOpen(testApp App) {
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
	Expect(strings.TrimSpace(string(output))).To(Equal(testApp.URL))
}
