package tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	ourSemVerRegExp := `^v?(?:[0-9]+\.){2}[0-9]+`
	gitSHARegExp := `[0-9a-f]{7}`

	It("prints its version", func() {
		output, err := execute("deis --version")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(MatchRegexp(`%s(-dev)?(-%s)?\n`, ourSemVerRegExp, gitSHARegExp))
	})
})
