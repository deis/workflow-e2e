package tests

import (
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO (bacongobbler): inspect kubectl for limits being applied to manifest
var _ = Describe("Limits", func() {
	Context("with a deployed app", func() {

		var testApp App
		once := &sync.Once{}

		BeforeEach(func() {
			// Set up the Limits test app only once and assume the suite will clean up.
			once.Do(func() {
				testApp = deployApp("example-go")
			})
		})

		It("can list limits", func() {
			sess, err := execute("deis limits:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess).To(SatisfyAll(
				ContainSubstring("=== %s Limits", testApp.Name),
				ContainSubstring("--- Memory\nUnlimited"),
				ContainSubstring("--- CPU\nUnlimited"),
			))
		})

		It("can set a memory limit", func() {
			sess, err := execute("deis limits:set cmd=64M -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess).To(ContainSubstring("--- Memory\ncmd     64M"))
			// Check that --memory also works too
			sess, err = execute("deis limits:set --memory cmd=128M -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess).To(ContainSubstring("--- Memory\ncmd     128M"))
		})

		It("can set a CPU limit", func() {
			sess, err := execute("deis limits:set --cpu cmd=1024 -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess).To(ContainSubstring("--- CPU\ncmd     1024"))
		})

		It("can unset a memory limit", func() {
			sess, err := execute("deis limits:unset cmd -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred(), sess)
			Expect(sess).To(ContainSubstring("--- Memory\nUnlimited"))

			// Check that --memory works too
			sess, err = execute("deis limits:set --memory cmd=64M -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred(), sess)
			Expect(sess).To(ContainSubstring("--- Memory\ncmd     64M"))
			sess, err = execute("deis limits:unset --memory cmd -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred(), sess)
			Expect(sess).To(ContainSubstring("--- Memory\nUnlimited"))
		})

		It("can unset a CPU limit", func() {
			sess, err := execute("deis limits:unset --cpu cmd -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred(), sess)
			Expect(sess).To(ContainSubstring("--- CPU\nUnlimited"))
		})
	})
})
