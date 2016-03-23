package tests

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// TODO (bacongobbler): inspect kubectl for limits being applied to manifest
var _ = Describe("Limits", func() {
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

		It("can list limits", func() {
			sess, err := start("deis limits:list -a %s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say(fmt.Sprintf("=== %s Limits", testApp.Name)))
			Eventually(sess).Should(Say("--- Memory\nUnlimited"))
			Eventually(sess).Should(Say("--- CPU\nUnlimited"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can set a memory limit", func() {
			sess, err := start("deis limits:set cmd=64M -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- Memory\ncmd     64M"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			// Check that --memory also works too
			sess, err = start("deis limits:set --memory cmd=128M -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- Memory\ncmd     128M"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can set a CPU limit", func() {
			sess, err := start("deis limits:set --cpu cmd=1024 -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- CPU\ncmd     1024"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can unset a memory limit", func() {
			sess, err := start("deis limits:unset cmd -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- Memory\nUnlimited"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			// Check that --memory works too
			sess, err = start("deis limits:set --memory cmd=64M -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- Memory\ncmd     64M"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis limits:unset --memory cmd -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- Memory\nUnlimited"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can unset a CPU limit", func() {
			sess, err := start("deis limits:unset --cpu cmd -a %s", testData.Profile, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("--- CPU\nUnlimited"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
