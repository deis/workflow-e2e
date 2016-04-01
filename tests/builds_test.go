package tests

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// createBuild invokes deis builds:create <image> -a <app>
// with provided <image> on provided <app>
// and validates that no errors have occurred and build was successful
func createBuild(profile string, image string, app App, options ...string) {
	pullOrCreateBuild(profile, image, app, "builds:create", strings.Join(options, " "))
}

// deisPull invokes deis pull <image> -a <app>
// with provided <image> on provided <app>
// and validates that no errors have occurred and build was successful
func deisPull(profile string, image string, app App, options ...string) {
	pullOrCreateBuild(profile, image, app, "pull", strings.Join(options, " "))
}

func pullOrCreateBuild(profile string, image string, app App, command string, options string) {
	sess, err := start("deis %s %s -a %s %s", profile, command, image, app.Name, options)
	Eventually(sess).Should(Say("Creating build..."))
	Eventually(sess, defaultMaxTimeout).Should(Say("done"))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("Builds", func() {
	Context("with a logged-in user", func() {
		var exampleRepo string
		var exampleImage string
		var testApp App
		var testData TestData

		BeforeEach(func() {
			testData = initTestData()
			exampleRepo = "example-go"
			exampleImage = fmt.Sprintf("deis/%s:latest", exampleRepo)
			testApp.Name = getRandAppName()
			gitInit()
		})

		AfterEach(func() {
			gitClean()
		})

		Context("with no app", func() {
			It("cannot create a build without existing app", func() {
				sess, err := start("deis builds:create %s -a %s", testData.Profile, exampleImage, testApp.Name)
				Eventually(sess.Err).Should(Say("404 Not Found"))
				Eventually(sess).Should(Exit(1))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with existing app", func() {
			var testData TestData

			BeforeEach(func() {
				testData = initTestData()
				createApp(testData.Profile, testApp.Name)
				createBuild(testData.Profile, exampleImage, testApp)
			})

			It("can list app builds", func() {
				sess, err := start("deis builds:list -a %s", testData.Profile, testApp.Name)
				Eventually(sess).Should(Say(uuidRegExp))
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with a deployed app", func() {
			var curlCmd Cmd
			var cmdRetryTimeout int
			var procFile string
			var testData TestData

			BeforeEach(func() {
				testData = initTestData()
				cmdRetryTimeout = 60
				procFile = fmt.Sprintf("worker: while true; do echo hi; sleep 3; done")
				testApp.URL = strings.Replace(getController(), "deis", testApp.Name, 1)
				createApp(testData.Profile, testApp.Name, "--no-remote")
				createBuild(testData.Profile, exampleImage, testApp)
			})

			It("can list app builds", func() {
				sess, err := start("deis builds:list -a %s", testData.Profile, testApp.Name)
				Eventually(sess).Should(Say(uuidRegExp))
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
			})

			It("can create a build from an existing image (\"deis pull\")", func() {
				procsListing := listProcs(testData.Profile, testApp).Out.Contents()
				// scrape current processes, should be 1 (cmd)
				Expect(len(scrapeProcs(testApp.Name, procsListing))).To(Equal(1))

				// curl app to make sure everything OK
				curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

				deisPull(testData.Profile, exampleImage, testApp, fmt.Sprintf(`--procfile="%s"`, procFile))

				sess, err := start("deis ps:scale worker=1 -a %s", testData.Profile, testApp.Name)
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())

				procsListing = listProcs(testData.Profile, testApp).Out.Contents()
				// scrape current processes, should be 2 (1 cmd, 1 worker)
				Expect(len(scrapeProcs(testApp.Name, procsListing))).To(Equal(2))

				// TODO: https://github.com/deis/workflow-e2e/issues/84
				// "deis logs -a %s", app
				// sess, err = start("deis logs -a %s", testData.Profile, testApp.Name)
				// Expect(err).To(BeNil())
				// Eventually(sess).Should(Say("hi"))
				// Eventually(sess).Should(Exit(0))
				// Expect(err).NotTo(HaveOccurred())

				// curl app to make sure everything OK
				curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

				// can scale cmd down to 0
				sess, err = start("deis ps:scale cmd=0 -a %s", testData.Profile, testApp.Name)
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())

				procsListing = listProcs(testData.Profile, testApp).Out.Contents()
				// scrape current processes, should be 1 worker
				Expect(len(scrapeProcs(testApp.Name, procsListing))).To(Equal(1))

				// TODO: still susceptible to intermittent curl timeouts
				// (not returning in under 10 seconds)
				// with routable 'cmd' process gone, curl should return StatusServiceUnavailable
				//curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				//Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusServiceUnavailable), cmdRetryTimeout)).Should(BeTrue())
			})
		})
	})
})
