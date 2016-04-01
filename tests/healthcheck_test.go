package tests

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Healthcheck", func() {
	var testData TestData
	appName := getRandAppName()

	Context("with a deployed app", func() {
		// create and deploy an app
		BeforeEach(func() {
			testData = initTestData()
			sess, err := start("deis apps:create %s", testData.Profile, appName)
			Eventually(sess).Should(Say("Creating Application... done, created %s", appName))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis pull deis/example-go -a %s", testData.Profile, appName)
			Eventually(sess).Should(Say("Creating build... done"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		// TODO: test is broken
		XIt("can stay running during a scale event", func() {
			router, err := getRawRouter()
			Expect(err).To(BeNil())
			appURLStr := fmt.Sprintf("%s://%s.%s", router.Scheme, appName, router.Host)
			stopCh := make(chan struct{})
			doneCh := make(chan struct{})

			// start scaling the app
			go func() {
				for range stopCh {
					sess, err := start("deis ps:scale web=4 -a %s", testData.Profile, appName)
					Eventually(sess).Should(Exit(0))
					Expect(err).NotTo(HaveOccurred())
				}
				close(doneCh)
			}()

			for i := 0; i < 10; i++ {
				// start the scale operation. waits until the last scale op has finished
				stopCh <- struct{}{}
				resp, err := http.Get(appURLStr)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
			}

			// wait until the goroutine that was scaling the app shuts down. not strictly necessary, just good practice
			Eventually(doneCh).Should(BeClosed())
		})

	})
})
