package tests

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var procsRegexp = `(%s-v\d+-[\w-]+) up \(v\d+\)`

// TODO: https://github.com/deis/workflow-e2e/issues/108
//       for example, these could live in common/certs.go
// certs-specific common actions and expectations
func listProcs(profile string, testApp App) *Session {
	sess, err := start("deis ps:list --app=%s", profile, testApp.Name)
	Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	return sess
}

// scrapeProcs returns the sorted process names for an app from the given output.
// It matches the current "deis ps" output for a healthy container:
//   earthy-vocalist-v2-cmd-1d73e up (v2)
//   myapp-v16-web-bujlq up (v16)
func scrapeProcs(app string, output []byte) []string {
	re := regexp.MustCompile(fmt.Sprintf(procsRegexp, app))
	found := re.FindAllSubmatch(output, -1)
	procs := make([]string, len(found))
	for i := range found {
		procs[i] = string(found[i][1])
	}
	sort.Strings(procs)
	return procs
}

var _ = Describe("Processes", func() {

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

		DescribeTable("can scale up and down",

			func(scaleTo, respCode int) {
				// TODO: need some way to choose between "web" and "cmd" here!
				// scale the app's processes to the desired number
				sess, err := start("deis ps:scale web=%d --app=%s", testData.Profile, scaleTo, testApp.Name)
				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// test that there are the right number of processes listed
				procsListing := listProcs(testData.Profile, testApp).Out.Contents()
				procs := scrapeProcs(testApp.Name, procsListing)
				Expect(len(procs)).To(Equal(scaleTo))

				// curl the app's root URL and print just the HTTP response code
				sess, err = start(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testData.Profile, testApp.URL)
				Eventually(sess).Should(Say(strconv.Itoa(respCode)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			},

			Entry("scales to 1", 1, 200),
			Entry("scales to 3", 3, 200),
			PEntry("scales to 0", 0, 503),
			PEntry("scales to 3", 3, 200),
			PEntry("scales to 0", 0, 503),
			PEntry("scales to 1", 1, 200),
		)

		DescribeTable("can restart processes",

			func(restart string, scaleTo int, respCode int) {
				// TODO: need some way to choose between "web" and "cmd" here!
				// scale the app's processes to the desired number
				sess, err := start("deis ps:scale web=%d --app=%s", testData.Profile, scaleTo, testApp.Name)

				Eventually(sess).Should(Say("Scaling processes... but first,"))
				Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
				Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// capture the process names
				beforeProcs := scrapeProcs(testApp.Name, sess.Out.Contents())

				// restart the app's process(es)
				var arg string
				switch restart {
				case "all":
					arg = ""
				case "by type":
					// TODO: need some way to choose between "web" and "cmd" here!
					arg = "web"
				case "by wrong type":
					// TODO: need some way to choose between "web" and "cmd" here!
					arg = "cmd"
				case "one":
					procsLen := len(beforeProcs)
					Expect(procsLen).To(BeNumerically(">", 0))
					arg = beforeProcs[rand.Intn(procsLen)]
				}
				sess, err = start("deis ps:restart %s --app=%s", testData.Profile, arg, testApp.Name)

				Eventually(sess).Should(Say("Restarting processes... but first,"))
				if scaleTo == 0 || restart == "by wrong type" {
					Eventually(sess, defaultMaxTimeout).Should(Say("Could not find any processes to restart"))
				} else {
					Eventually(sess, defaultMaxTimeout).Should(Say(`done in \d+s`))
					Eventually(sess).Should(Say("=== %s Processes", testApp.Name))
				}
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// capture the process names
				procsListing := listProcs(testData.Profile, testApp).Out.Contents()
				afterProcs := scrapeProcs(testApp.Name, procsListing)

				// compare the before and after sets of process names
				Expect(len(afterProcs)).To(Equal(scaleTo))
				if scaleTo > 0 && restart != "by wrong type" {
					Expect(beforeProcs).NotTo(Equal(afterProcs))
				}

				// curl the app's root URL and print just the HTTP response code
				cmdRetryTimeout := 60
				curlCmd := Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
				Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
			},

			Entry("restarts one of 1", "one", 1, 200),
			Entry("restarts all of 1", "all", 1, 200),
			Entry("restarts all of 1 by type", "by type", 1, 200),
			Entry("restarts all of 1 by wrong type", "by wrong type", 1, 200),
			Entry("restarts one of 3", "one", 3, 200),
			Entry("restarts all of 3", "all", 3, 200),
			Entry("restarts all of 3 by type", "by type", 3, 200),
			Entry("restarts all of 3 by wrong type", "by wrong type", 3, 200),
			PEntry("restarts all of 0", "all", 0, 503),
			PEntry("restarts all of 0 by type", "by type", 0, 503),
			PEntry("restarts all of 0 by wrong type", "by wrong type", 0, 503),
		)
	})
})
