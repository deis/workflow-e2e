package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Perms", func() {
	AfterEach(func() {
		gitClean()
	})

	Context("when logged in as a normal user", func() {
		var testData TestData
		var testData2 TestData
		var testApp App

		BeforeEach(func() {
			testData = initTestData()
			testData2 = initTestData()
			testApp.Name = getRandAppName()
			gitInit()
			createApp(testData.Profile, testApp.Name)
		})

		It("can't create, list, or delete admin permissions", func() {
			sess, err := start("deis perms:create %s --admin", testData.Profile, adminTestData.Username)
			Eventually(sess, defaultMaxTimeout).Should(Say("Adding admin to system administrators..."))
			Eventually(sess.Err).Should(Say("403 Forbidden"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:list --admin", testData.Profile)
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("403 Forbidden"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:delete %s --admin", testData.Profile, adminTestData.Username)
			Eventually(sess.Err, defaultMaxTimeout).Should(Say("403 Forbidden"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
			sess, err = start("deis perms:list --admin", testData.Profile)
			Eventually(sess.Err).Should(Say("403 Forbidden"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can create, list, and delete app permissions", func() {
			sess, err := start("deis perms:create %s --app=%s", testData.Profile, testData2.Username, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("Adding %s to %s collaborators... done\n", testData2.Username, testApp.Name))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis perms:list --app=%s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).Should(Say("%s", testData2.Username))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis perms:delete %s --app=%s", testData.Profile, testData2.Username, testApp.Name)
			Eventually(sess, defaultMaxTimeout).Should(Say("Removing %s from %s collaborators... done", testData2.Username, testApp.Name))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis perms:list --app=%s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).ShouldNot(Say("%s", testData2.Username))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
