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
func createBuild(image string, app App, options ...string) {
	pullOrCreateBuild(image, app, "builds:create", strings.Join(options, " "))
}

// deisPull invokes deis pull <image> -a <app>
// with provided <image> on provided <app>
// and validates that no errors have occurred and build was successful
func deisPull(image string, app App, options ...string) {
	pullOrCreateBuild(image, app, "pull", strings.Join(options, " "))
}

func pullOrCreateBuild(image string, app App, command string, options string) {
	sess, err := start("deis %s %s -a %s %s", command, image, app.Name, options)
	Expect(err).To(BeNil())
	Eventually(sess, defaultMaxTimeout).Should(Exit(0))
	Eventually(sess).Should(Say("Creating build..."))
	Eventually(sess).Should(Say("done"))
}

var _ = Describe("Builds", func() {
	Context("with a logged-in user", func() {
		var exampleRepo string
		var exampleImage string
		var testApp App

		BeforeEach(func() {
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
				cmd, err := start("deis builds:create %s -a %s", exampleImage, testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(cmd.Err).Should(Say("404 Not Found"))
				Eventually(cmd).Should(Exit(1))
			})
		})

		Context("with existing app", func() {

			BeforeEach(func() {
				createApp(testApp.Name)
				createBuild(exampleImage, testApp)
			})

			It("can list app builds", func() {
				cmd, err := start("deis builds:list -a %s", testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(cmd).Should(Exit(0))
				Eventually(cmd).Should(Say(uuidRegExp))
			})
		})

		Context("with a deployed app", func() {
			var curlCmd Cmd
			var cmdRetryTimeout int
			var procFile string

			BeforeEach(func() {
				cmdRetryTimeout = 10
				procFile = fmt.Sprintf("worker: while true; do echo hi; sleep 3; done")
				testApp.URL = strings.Replace(url, "deis", testApp.Name, 1)
				createApp(testApp.Name, "--no-remote")
				createBuild(exampleImage, testApp)
			})

			It("can list app builds", func() {
				cmd, err := start("deis builds:list -a %s", testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(cmd).Should(Exit(0))
				Eventually(cmd).Should(Say(uuidRegExp))
			})

			PIt("can create a build from an existing image (\"deis pull\")", func() {
				procsListing := listProcs(testApp).Out.Contents()
				// scrape current processes, should be 1 (cmd)
				Expect(len(scrapeProcs(testApp.Name, procsListing))).To(Equal(1))

				// curl app to make sure everything OK
				curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

				deisPull(exampleImage, testApp, fmt.Sprintf(`--procfile="%s"`, procFile))

				sess, err := start("deis ps:scale worker=1 -a %s", testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Eventually(sess).Should(Exit(0))

				procsListing = listProcs(testApp).Out.Contents()
				// scrape current processes, should be 2 (1 cmd, 1 worker)
				Expect(len(scrapeProcs(testApp.Name, procsListing))).To(Equal(2))

				// TODO: https://github.com/deis/workflow-e2e/issues/84
				// "deis logs -a %s", app
				// sess, err = start("deis logs -a %s", testApp.Name)
				// Expect(err).To(BeNil())
				// Eventually(sess).Should(Say("hi"))
				// Eventually(sess).Should(Exit(0))

				// curl app to make sure everything OK
				curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

				// can scale cmd down to 0
				sess, err = start("deis ps:scale cmd=0 -a %s", testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Eventually(sess).Should(Exit(0))

				procsListing = listProcs(testApp).Out.Contents()
				// scrape current processes, should be 1 worker
				Expect(len(scrapeProcs(testApp.Name, procsListing))).To(Equal(1))

				// with routable 'cmd' process gone, curl should return StatusBadGateway
				curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusBadGateway), cmdRetryTimeout)).Should(BeTrue())
			})
		})
	})
})
