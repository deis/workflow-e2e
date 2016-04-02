package tests

import (
	"github.com/deis/workflow-e2e/tests/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("deis help", func() {

	usageMsg := "Usage: deis <command> [<args>...]"
	optionFlagsMsg := "Option flags::"
	noMatchMsg := "Found no matching command, try 'deis help'"

	Specify("the --help flag causes the help message to be printed", func() {
		output, err := cmd.Execute("deis --help")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(usageMsg))
		Expect(output).To(ContainSubstring(optionFlagsMsg))
	})

	Specify("the -h flag causes the help message to be printed", func() {
		output, err := cmd.Execute("deis -h")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(usageMsg))
		Expect(output).To(ContainSubstring(optionFlagsMsg))
	})

	Specify("the help subcommand causes the help message to be printed", func() {
		output, err := cmd.Execute("deis help")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(usageMsg))
		Expect(output).To(ContainSubstring(optionFlagsMsg))
	})

	Specify("deis invoked with no flags or subcommands causes the usage message to be printed", func() {
		output, err := cmd.Execute("deis")
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring(usageMsg))
		Expect(output).NotTo(ContainSubstring(optionFlagsMsg))
	})

	Specify("an invalid subcommand causes an error message to be printed", func() {
		output, err := cmd.Execute("deis bogus-command")
		Expect(err).To(HaveOccurred())
		Expect(output).To(SatisfyAll(
			ContainSubstring(noMatchMsg),
			ContainSubstring(usageMsg)))
	})

})
