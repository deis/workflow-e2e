package tests

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Releases", func() {
	Context("with a deployed app", func() {
		var appName string

		BeforeEach(func() {
			appName = getRandAppName()
			cmd := createApp(appName)
			Eventually(cmd).Should(SatisfyAll(
				Say("Git remote deis added"),
				Say("remote available at ")))
		})

		// 500's everytime
		XIt("can deploy the app", func() {
			sess, err := start("deis pull deis/example-go -a %s", appName)
			Expect(err).To(BeNil())
			Eventually(sess, (10 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say("Creating build... done"))
		})

		It("can list releases", func() {
			sess, err := start("deis releases:list -a %s", appName)
			Expect(err).To(BeNil())
			Eventually(sess, (1 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say("=== %s Releases", appName))
			Eventually(sess).Should(Say(`v1\s+.*\s+%s created initial release`, testUser))
		})

		It("can rollback to a previous release", func() {
			sess, err := start("deis releases:rollback v1 -a %s", appName)
			Expect(err).To(BeNil())
			Eventually(sess, (1 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say(`Rolling back to`))
			Eventually(sess).Should(Say(`...done`))
		})

		It("can get info on releases", func() {
			sess, err := start("deis releases:info v1 -a %s", appName)
			Expect(err).To(BeNil())
			Eventually(sess, (1 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say("=== %s Release v1", appName))
			Eventually(sess).Should(Say(`config:\s+[\w-]+`))
			Eventually(sess).Should(Say(`owner:\s+%s`, testUser))
			Eventually(sess).Should(Say(`summary:\s+%s \w+`, testUser))
			// the below updated date has to match a string like 2015-12-22T21:20:31UTC
			Eventually(sess).Should(Say(`updated:\s+[\w\-\:]+UTC`))
			Eventually(sess).Should(Say(`uuid:\s+[0-9a-f\-]+`))
		})
	})
})
