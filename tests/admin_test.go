package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Admin", func() {
	Context("when logged in as an admin user", func() {
		var testData TestData
		var testApp App

		BeforeEach(func() {
			testData = initTestData()
			testApp.Name = getRandAppName()
			gitInit()
			createApp(testData.Profile, testApp.Name)
		})

		It("can create, list, and delete admin permissions", func() {
			sess, err := start("deis perms:create %s --admin", adminTestData.Profile, testData.Username)
			Eventually(sess, defaultMaxTimeout).Should(Say("Adding %s to system administrators... done\n", testData.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:list --admin", adminTestData.Profile)
			Eventually(sess).Should(Say("=== Administrators"))
			Eventually(sess).Should(Say(adminTestData.Username))
			Eventually(sess).Should(Say(testData.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:delete %s --admin", adminTestData.Profile, testData.Username)
			Eventually(sess, defaultMaxTimeout).Should(Say("Removing %s from system administrators... done", testData.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:list --admin", adminTestData.Profile)
			Eventually(sess).Should(Say("=== Administrators"))
			Expect(sess).ShouldNot(Say(testData.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can create, list, and delete app permissions", func() {
			sess, err := start("deis perms:create %s --app=%s", adminTestData.Profile, testData.Username, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("Adding %s to %s collaborators... done\n", testData.Username, testApp.Name))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:list --app=%s", adminTestData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).Should(Say("%s", testData.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:delete %s --app=%s", adminTestData.Profile, testData.Username, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("Removing %s from %s collaborators... done", testData.Username, testApp.Name))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:list --app=%s", adminTestData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).ShouldNot(Say("%s", testData.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("regenerates the token for a specified user", func() {
			sess, err := start("deis auth:regenerate -u %s", adminTestData.Profile, testData.Username)
			Eventually(sess).Should(Say("Token Regenerated"))
			Expect(err).NotTo(HaveOccurred())
		})

		// This is marked pending because it resets all user auth tokens. Because we run the tests in parallel
		// this can wreak havoc on tests that may be in flight. We will need to reevaluate how we want to test this functionality.
		XIt("regenerates the token for all users", func() {
			sess, err := start("deis auth:regenerate --all", adminTestData.Profile)
			Eventually(sess).Should(Say("Token Regenerated"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can list all users", func() {
			sess, err := start("deis users:list", adminTestData.Profile)
			Eventually(sess).Should(Say("=== Users"))
			output := string(sess.Out.Contents())
			Expect(output).To(ContainSubstring(adminTestData.Username))
			Expect(output).To(ContainSubstring(testData.Username))
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
