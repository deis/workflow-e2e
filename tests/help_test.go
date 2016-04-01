package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const noMatch string = "Found no matching command, try 'deis help'"
const usage string = "Usage: deis <command> [<args>...]"

var _ = Describe("Help", func() {

	It("prints help on --help", func() {
		output, err := execute("deis %s", "--help")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(usage))
	})

	It("prints help on -h", func() {
		output, err := execute("deis %s", "-h")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(usage))
	})

	It("prints help on help", func() {
		output, err := execute("deis %s", "help")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(usage))
	})

	It("defaults to a usage message", func() {
		output, err := execute("deis")
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring(usage))
	})

	It("rejects a bogus command", func() {
		output, err := execute("deis bogus-command")
		Expect(err).To(HaveOccurred())
		Expect(output).To(SatisfyAll(
			ContainSubstring(noMatch),
			ContainSubstring(usage)))
	})
})
