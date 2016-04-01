package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Releases", func() {
	var testApp App
	var exampleImage string
	var testData TestData

	Context("with a deployed app", func() {
		exampleImage = "deis/example-go"

		BeforeEach(func() {
			testData = initTestData()
			gitInit()
			testApp = App{Name: getRandAppName()}
			createApp(testData.Profile, testApp.Name)
		})

		It("can deploy the app", func() {
			// generate v2 release
			deisPull(testData.Profile, exampleImage, testApp)

			// "can list releases"
			sess, err := start("deis releases:list -a %s", testData.Profile, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Releases", testApp.Name))
			Eventually(sess).Should(Say(`v1\s+.*\s+%s created initial release`, testData.Username))
			Eventually(sess).Should(Exit(0))

			// generate v3 release
			deisPull(testData.Profile, exampleImage, testApp)

			// "can rollback to a previous release"
			sess, err = start("deis releases:rollback v2 -a %s", testData.Profile, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say(`Rolling back to`))
			Eventually(sess, defaultMaxTimeout).Should(Say(`...done`))
			Eventually(sess).Should(Exit(0))

			// "can get info on releases"
			sess, err = start("deis releases:info v2 -a %s", testData.Profile, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Release v2", testApp.Name))
			Eventually(sess).Should(Say(`config:\s+[\w-]+`))
			Eventually(sess).Should(Say(`owner:\s+%s`, testData.Username))
			Eventually(sess).Should(Say(`summary:\s+%s \w+`, testData.Username))
			// the below updated date has to match a string like 2015-12-22T21:20:31UTC
			Eventually(sess).Should(Say(`updated:\s+[\w\-\:]+UTC`))
			Eventually(sess).Should(Say(`uuid:\s+[0-9a-f\-]+`))
			Eventually(sess).Should(Exit(0))

			//TODO: add actions/validations around scenario described in
			// https://github.com/deis/controller/issues/540
		})
	})
})
