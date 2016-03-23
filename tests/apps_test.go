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

	Context("with no app", func() {
		var testApp App
		var testData TestData

		BeforeEach(func() {
			testApp.Name = getRandAppName()
			testData = initTestData()
		})

		It("can't get app info", func() {
			sess, err := start("deis info -a %s", testData.Profile, testApp.Name)
			Eventually(sess.Err).Should(Say("Not found."))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can't get app logs", func() {
			sess, err := start("deis logs -a %s", testData.Profile, testApp.Name)
			Eventually(sess.Err).Should(Say(`Error: There are currently no log messages. Please check the following things:`))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can't run a command in the app environment", func() {
			sess, err := start("deis apps:run echo Hello, 世界", testData.Profile)
			Eventually(sess).Should(Say("Running 'echo Hello, 世界'..."))
			Eventually(sess.Err).Should(Say("Not found."))
			Eventually(sess).ShouldNot(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can't open a bogus app URL", func() {
			sess, err := start("deis open -a %s", testData.Profile, getRandAppName())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})

	})

	Context("when creating an app", func() {
		var testApp App
		var testData TestData

		BeforeEach(func() {
			testData = initTestData()
			testApp.Name = getRandAppName()
			gitInit()
		})

		It("creates an app with a git remote", func() {
			sess, err := start("deis apps:create %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("created %s", testApp.Name))
			Eventually(sess).Should(Say(`Git remote deis added`))
			Eventually(sess).Should(Say(`remote available at `))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates an app with no git remote", func() {
			sess, err := start("deis apps:create %s --no-remote", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("created %s", testApp.Name))
			Eventually(sess).Should(Say("remote available at "))
			Eventually(sess).ShouldNot(Say("Git remote deis added"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates an app with a custom buildpack", func() {
			sess, err := start("deis apps:create %s --buildpack https://example.com", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("created %s", testApp.Name))
			Eventually(sess).Should(Say("Git remote deis added"))
			Eventually(sess).Should(Say("remote available at "))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis config:list -a %s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("BUILDPACK_URL"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with a deployed app", func() {
		var testApp App
		var testData TestData

		BeforeEach(func() {
			testData = initTestData()
			os.Chdir("example-go")
			appName := getRandAppName()
			createApp(testData.Profile, appName)
			testApp = deployApp(testData.Profile, appName)
		})

		AfterEach(func() {
			defer os.Chdir("..")
		})

		It("can't create an existing app", func() {
			sess, err := start("deis apps:create %s", testData.Profile, testApp.Name)
			Eventually(sess.Err).Should(Say("App with this id already exists."))
			Eventually(sess).ShouldNot(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can get app info", func() {
			verifyAppInfo(testData.Profile, testData.Username, testApp.Name, testApp.URL)
		})

		// V broken
		XIt("can get app logs", func() {
			sess, err := start("deis logs", testData.Profile)
			Eventually(sess).Should(SatisfyAll(
				Say("%s\\[deis-controller\\]\\: %s created initial release", testApp.Name, testData.Username),
				Say("%s\\[deis-controller\\]\\: %s deployed", testApp.Name, testData.Username),
				Say("%s\\[deis-controller\\]\\: %s scaled containers", testApp.Name, testData.Username)))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can open the app's URL", func() {
			verifyAppOpen(testData.Profile, testApp.URL)
		})

		It("can run a command in the app environment", func() {
			sess, err := start("deis apps:run echo Hello, 世界", testData.Profile)
			Eventually(sess, (1 * time.Minute)).Should(Say("Hello, 世界"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can transfer the app to another owner", func() {
			sess, err := start("deis apps:transfer %s", testData.Profile, adminTestData.Username)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			sess, err = start("deis info -a %s", testData.Profile, testApp.Name)
			Eventually(sess.Err).Should(Say("You do not have permission to perform this action."))
			Eventually(sess).Should(Exit(1))
		})
	})

	Context("with a custom buildpack deployed app", func() {
		var testApp App
		var testData TestData

		BeforeEach(func() {
			testData = initTestData()
			os.Chdir("example-perl")
			appName := getRandAppName()
			createApp(testData.Profile, appName, "--buildpack", "https://github.com/miyagawa/heroku-buildpack-perl.git")
			testApp = deployApp(testData.Profile, appName)
		})

		It("can get app info", func() {
			verifyAppInfo(testData.Profile, testData.Username, testApp.Name, testApp.URL)
		})

		It("can open the app's URL", func() {
			verifyAppOpen(testData.Profile, testApp.URL)
		})

	})
})

func verifyAppInfo(profile string, username string, appName string, url string) {
	sess, err := start("deis info -a %s", profile, appName)
	Eventually(sess).Should(Say("=== %s Application", appName))
	Eventually(sess).Should(Say(`uuid:\s*%s`, uuidRegExp))
	Eventually(sess).Should(Say(`url:\s*%s`, strings.Replace(url, "http://", "", 1)))
	Eventually(sess).Should(Say(`owner:\s*%s`, username))
	Eventually(sess).Should(Say(`id:\s*%s`, appName))
	Eventually(sess).Should(Say("=== %s Processes", appName))
	Eventually(sess).Should(Say(procsRegexp, appName))
	Eventually(sess).Should(Say("=== %s Domains", appName))
	Eventually(sess).Should(Say("%s", appName))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

func verifyAppOpen(profile string, url string) {
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
	sess, err := startCmd(Cmd{Env: env, CommandLineString: "DEIS_PROFILE=" + profile + " deis open"})
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())

	// check shim output
	output, err := ioutil.ReadFile(myShim.OutFile.Name())
	Expect(err).NotTo(HaveOccurred())
	Expect(strings.TrimSpace(string(output))).To(Equal(url))
}
