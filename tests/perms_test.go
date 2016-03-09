package tests

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Perms", func() {
	var testApp App

	BeforeEach(func() {
		os.Chdir("example-go")
		appName := getRandAppName()
		createApp(appName)
		testApp = deployApp(appName)
	})

	AfterEach(func() {
		defer os.Chdir("..")
		destroyApp(testApp)
	})

	Context("when logged in as an admin user", func() {
		BeforeEach(func() {
			login(url, testAdminUser, testAdminPassword)
		})

		It("can create, list, and delete admin permissions", func() {
			output, err := execute("deis perms:create %s --admin", testUser)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(
				ContainSubstring("Adding %s to system administrators... done\n", testUser))
			output, err = execute("deis perms:list --admin")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(SatisfyAll(
				HavePrefix("=== Administrators"),
				ContainSubstring(testUser),
				ContainSubstring(testAdminUser)))
			output, err = execute("deis perms:delete %s --admin", testUser)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(
				ContainSubstring("Removing %s from system administrators... done", testUser))
			output, err = execute("deis perms:list --admin")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(SatisfyAll(
				HavePrefix("=== Administrators"),
				ContainSubstring(testAdminUser)))
			Expect(output).NotTo(ContainSubstring(testUser))
		})

		It("can create, list, and delete app permissions", func() {
			sess, err := start("deis perms:create %s --app=%s", testUser, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Adding %s to %s collaborators... done\n", testUser, testApp.Name))

			sess, err = start("deis perms:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).Should(Say("%s", testUser))

			sess, err = start("deis perms:delete %s --app=%s", testUser, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Removing %s from %s collaborators... done", testUser, testApp.Name))

			sess, err = start("deis perms:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).ShouldNot(Say("%s", testUser))

			Eventually(sess).Should(Exit(0))
		})
	})

	Context("when logged in as a normal user", func() {
		It("can't create, list, or delete admin permissions", func() {
			output, err := execute("deis perms:create %s --admin", testAdminUser)
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("403 Forbidden"))
			output, err = execute("deis perms:list --admin")
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("403 Forbidden"))
			output, err = execute("deis perms:delete %s --admin", testAdminUser)
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("403 Forbidden"))
			output, err = execute("deis perms:list --admin")
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("403 Forbidden"))
		})

		It("can create, list, and delete app permissions", func() {
			sess, err := start("deis perms:create %s --app=%s", testUser, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Adding %s to %s collaborators... done\n", testUser, testApp.Name))

			sess, err = start("deis perms:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).Should(Say("%s", testUser))

			sess, err = start("deis perms:delete %s --app=%s", testUser, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Removing %s from %s collaborators... done", testUser, testApp.Name))

			sess, err = start("deis perms:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s's Users", testApp.Name))
			Eventually(sess).ShouldNot(Say("%s", testUser))

			Eventually(sess).Should(Exit(0))
		})
	})
})
