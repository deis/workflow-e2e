package tests

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// createBuild invokes deis builds:create <image> -a <app>
// with provided <image> on provided <app>
// and validates that no errors have occurred and build was successful
func createBuild(image string, testApp App) {
	cmd, err := start("deis builds:create %s -a %s", image, testApp.Name)
	Expect(err).NotTo(HaveOccurred())
	Eventually(cmd, defaultMaxTimeout).Should(Exit(0))
	Eventually(cmd).Should(Say("Creating build..."))
	Eventually(cmd).Should(Say("done"))
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
				gitInit()
				createApp(testApp.Name)
				createBuild(exampleImage, testApp)
			})

			AfterEach(func() {
				destroyApp(testApp)
				gitClean()
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
				cmd := createApp(testApp.Name, "--no-remote")
				Eventually(cmd).Should(Not(Say("Git remote deis added")))

				os.Chdir(exampleRepo)
				appName := getRandAppName()
				createApp(appName)
				testApp = deployApp(appName)
				procFile = fmt.Sprintf("worker: while true; do echo hi; sleep 3; done")

				createBuild(exampleImage, testApp)
			})

			AfterEach(func() {
				defer os.Chdir("..")
				destroyApp(testApp)
				gitClean()
			})

			It("can list app builds", func() {
				cmd, err := start("deis builds:list -a %s", testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(cmd).Should(Exit(0))
				Eventually(cmd).Should(Say(uuidRegExp))
			})

			It("can create a build from an existing image (\"deis pull\")", func() {
				// curl app to make sure everything OK
				curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

				sess, err := start(`deis pull %s -a %s --procfile="%s"`, exampleImage, testApp.Name, procFile)
				Expect(err).To(BeNil())
				Eventually(sess, defaultMaxTimeout).Should(Exit(0))
				Eventually(sess).Should(Say("Creating build..."))
				Eventually(sess).Should(Say("done"))

				sess, err = start("deis ps:scale worker=1 -a %s", testApp.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Eventually(sess).Should(Exit(0))

				// TODO: #84
				// "deis logs -a %s", app
				// sess, err = start("deis logs -a %s", testApp.Name)
				// Expect(err).To(BeNil())
				// Eventually(sess).Should(Say("hi"))
				// Eventually(sess).Should(Exit(0))
			})
		})
	})
})
