package tests

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Releases", func() {
	var testApp App
	var exampleImage string

	Context("with a deployed app", func() {
		exampleImage = "deis/example-go"

		BeforeEach(func() {
			url, testUser, testPassword, testEmail, keyName = createRandomUser()
			gitInit()
			testApp = App{Name: getRandAppName()}
			createApp(testApp.Name)
		})

		It("can deploy the app", func() {
			// generate v2 release
			deisPull(exampleImage, testApp)

			// "can list releases"
			sess, err := start("deis releases:list -a %s", testApp.Name)
			Expect(err).To(BeNil())
			Eventually(sess, (1 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say("=== %s Releases", testApp.Name))
			Eventually(sess).Should(Say(`v1\s+.*\s+%s created initial release`, testUser))

			// generate v3 release
			deisPull(exampleImage, testApp)

			// "can rollback to a previous release"
			sess, err = start("deis releases:rollback v2 -a %s", testApp.Name)
			Expect(err).To(BeNil())
			Eventually(sess, (1 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say(`Rolling back to`))
			Eventually(sess).Should(Say(`...done`))

			// "can get info on releases"
			sess, err = start("deis releases:info v2 -a %s", testApp.Name)
			Expect(err).To(BeNil())
			Eventually(sess, (1 * time.Minute)).Should(Exit(0))
			Eventually(sess).Should(Say("=== %s Release v2", testApp.Name))
			Eventually(sess).Should(Say(`config:\s+[\w-]+`))
			Eventually(sess).Should(Say(`owner:\s+%s`, testUser))
			Eventually(sess).Should(Say(`summary:\s+%s \w+`, testUser))
			// the below updated date has to match a string like 2015-12-22T21:20:31UTC
			Eventually(sess).Should(Say(`updated:\s+[\w\-\:]+UTC`))
			Eventually(sess).Should(Say(`uuid:\s+[0-9a-f\-]+`))

			//TODO: add actions/validations around scenario described in
			// https://github.com/deis/controller/issues/540
		})
	})
})
