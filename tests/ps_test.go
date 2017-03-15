package tests

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

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

var _ = Describe("deis ps", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.RegisterAndLogin()
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

			DescribeTable("that user can scale that app up and down",
				func(scaleTo, respCode int) {
					sess, err := cmd.Start("deis ps:scale cmd=%d --app=%s", &user, scaleTo, app.Name)
					Eventually(sess).Should(Say("Scaling processes... but first,"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`done in \d+s`))
					Eventually(sess).Should(Say("=== %s Processes", app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					// test that there are the right number of processes listed
					procsListing := listProcs(user, app, "").Out.Contents()
					procs := scrapeProcs(app.Name, procsListing)
					Expect(procs).To(HaveLen(scaleTo))

					// curl the app's root URL and print just the HTTP response code
					cmdRetryTimeout := 60
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(respCode), cmdRetryTimeout)).Should(BeTrue())
				},
				Entry("scales to 1", 1, 200),
				Entry("scales to 3", 3, 200),
				Entry("scales to 0", 0, 503),
			)

			DescribeTable("that user can interrupt a scaling event",
				func(scaleTo, respCode int) {

					sess, err := cmd.Start("deis ps:scale cmd=%d --app=%s", &user, scaleTo, app.Name)
					Eventually(sess).Should(Say("Scaling processes... but first,"))

					Expect(err).NotTo(HaveOccurred())

					// Sleep for a split second to ensure scale command makes it to the server.
					time.Sleep(200 * time.Millisecond)

					// Interrupt and wait for exit.
					sess = sess.Interrupt().Wait()

					// Ensure the right number of processes listed.
					Eventually(func() []string {
						procsListing := listProcs(user, app, "").Out.Contents()
						return scrapeProcs(app.Name, procsListing)
					}, settings.MaxEventuallyTimeout).Should(HaveLen(scaleTo))

					// curl the app's root URL and print just the HTTP response code
					cmdRetryTimeout := 60
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(respCode), cmdRetryTimeout)).Should(BeTrue())
				},
				Entry("scales to 3", 3, 200),
				Entry("scales to 0", 0, 503),
			)

			// TODO: Test is broken
			XIt("that app remains responsive during a scaling event", func() {
				stopCh := make(chan struct{})
				doneCh := make(chan struct{})

				// start scaling the app
				go func() {
					for range stopCh {
						sess, err := cmd.Start("deis ps:scale web=4 -a %s", &user, app.Name)
						Eventually(sess).Should(Exit(0))
						Expect(err).NotTo(HaveOccurred())
					}
					close(doneCh)
				}()

				for i := 0; i < 10; i++ {
					// start the scale operation. waits until the last scale op has finished
					stopCh <- struct{}{}
					resp, err := http.Get(app.URL)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
				}

				// wait until the goroutine that was scaling the app shuts down. not strictly necessary, just good practice
				Eventually(doneCh).Should(BeClosed())
			})

			DescribeTable("that user can restart that app's processes",
				func(restart string, scaleTo int, respCode int) {
					// TODO: need some way to choose between "web" and "cmd" here!
					// scale the app's processes to the desired number
					sess, err := cmd.Start("deis ps:scale cmd=%d --app=%s", &user, scaleTo, app.Name)

					Eventually(sess).Should(Say("Scaling processes... but first,"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`done in \d+s`))
					Eventually(sess).Should(Say("=== %s Processes", app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					// capture the process names
					beforeProcs := scrapeProcs(app.Name, sess.Out.Contents())

					// restart the app's process(es)
					var arg string
					switch restart {
					case "all":
						arg = ""
					case "by type":
						// TODO: need some way to choose between "web" and "cmd" here!
						arg = "cmd"
					case "by wrong type":
						// TODO: need some way to choose between "web" and "cmd" here!
						arg = "web"
					case "one":
						procsLen := len(beforeProcs)
						Expect(procsLen).To(BeNumerically(">", 0))
						arg = beforeProcs[rand.Intn(procsLen)]
					}
					sess, err = cmd.Start("deis ps:restart %s --app=%s", &user, arg, app.Name)
					Eventually(sess).Should(Say("Restarting processes... but first,"))
					if scaleTo == 0 || restart == "by wrong type" {
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Could not find any processes to restart"))
					} else {
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`done in \d+s`))
						Eventually(sess).Should(Say("=== %s Processes", app.Name))
					}
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					// capture the process names
					procsListing := listProcs(user, app, "").Out.Contents()
					afterProcs := scrapeProcs(app.Name, procsListing)

					// compare the before and after sets of process names
					Expect(afterProcs).To(HaveLen(scaleTo))
					if scaleTo > 0 && restart != "by wrong type" {
						Expect(beforeProcs).NotTo(Equal(afterProcs))
					}

					// curl the app's root URL and print just the HTTP response code
					cmdRetryTimeout := 60
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(respCode), cmdRetryTimeout)).Should(BeTrue())
				},
				Entry("restarts one of 1", "one", 1, 200),
				Entry("restarts all of 1", "all", 1, 200),
				Entry("restarts all of 1 by type", "by type", 1, 200),
				Entry("restarts all of 1 by wrong type", "by wrong type", 1, 200),
				Entry("restarts one of 3", "one", 3, 200),
				Entry("restarts all of 3", "all", 3, 200),
				Entry("restarts all of 3 by type", "by type", 3, 200),
				Entry("restarts all of 3 by wrong type", "by wrong type", 3, 200),
				Entry("restarts all of 0", "all", 0, 503),
				Entry("restarts all of 0 by type", "by type", 0, 503),
				Entry("restarts all of 0 by wrong type", "by wrong type", 0, 503),
			)

		})

	})

})

func listProcs(user model.User, app model.App, proctype string) *Session {
	sess, err := cmd.Start("deis ps:list --app=%s", &user, app.Name)
	Eventually(sess).Should(Say("=== %s Processes", app.Name))
	if proctype != "" {
		Eventually(sess).Should(Say("--- %s:", proctype))
	}
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	return sess
}

// scrapeProcs returns the sorted process names for an app from the given output.
// It matches the current "deis ps" output for a healthy container:
//   earthy-vocalist-cmd-123456789-1d73e up (v2)
//   myapp-web-123456789-bujlq up (v16)
func scrapeProcs(app string, output []byte) []string {
	procsRegexp := `(%s-[\w-]+) up \(v\d+\)`
	re := regexp.MustCompile(fmt.Sprintf(procsRegexp, app))
	found := re.FindAllSubmatch(output, -1)
	procs := make([]string, len(found))
	for i := range found {
		procs[i] = string(found[i][1])
	}
	sort.Strings(procs)
	return procs
}
