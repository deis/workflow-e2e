package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Auth", func() {
	Context("when logged out", func() {
		BeforeEach(func() {
			logout()
		})

		It("won't print the current user", func() {
			sess, err := start("deis auth:whoami", "")
			Eventually(sess.Err).Should(Say("Not logged in"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when logged in", func() {
		var testData TestData
		BeforeEach(func() {
			testData = initTestData()
		})

		It("can log out", func() {
			logout()
		})

		It("won't register twice", func() {
			cmd := "deis register %s --username=%s --password=%s --email=%s"
			sess, err := start(cmd, testData.Profile, testData.ControllerURL, testData.Username, testData.Password, testData.Email)
			Eventually(sess.Err).Should(Say("Registration failed"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})

		It("prints the current user", func() {
			sess, err := start("deis auth:whoami", testData.Profile)
			Eventually(sess).Should(Say("You are %s", testData.Username))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("regenerates the token for the current user", func() {
			sess, err := start("deis auth:regenerate", testData.Profile)
			Eventually(sess).Should(Say("Token Regenerated"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
