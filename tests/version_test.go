package tests

import (
	"github.com/deis/workflow-e2e/tests/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("deis version", func() {

	semVerRegExp := `^v?(?:[0-9]+\.){2}[0-9]+`
	gitSHARegExp := `[0-9a-f]{7}`

	Specify("the version subcommand causes version information to be printed", func() {
		output, err := cmd.Execute("deis --version")
		Expect(output).To(MatchRegexp(`%s(-dev)?(-%s)?\n`, semVerRegExp, gitSHARegExp))
		Expect(err).NotTo(HaveOccurred())
	})

})
