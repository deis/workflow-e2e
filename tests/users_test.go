package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Users", func() {
	Context("when logged in as a normal user", func() {
		var testData TestData

		BeforeEach(func() {
			testData = initTestData()
		})

		It("can't list all users", func() {
			sess, err := start("deis users:list", testData.Profile)
			Eventually(sess.Err, defaultMaxTimeout).Should(Say("403 Forbidden"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
