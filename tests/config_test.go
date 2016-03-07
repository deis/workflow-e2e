package tests

import (
	"io/ioutil"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Config", func() {

	Context("with a deployed app", func() {

		var testApp App
		once := &sync.Once{}

		BeforeEach(func() {
			// Set up the Processes test app only once and assume the suite will clean up.
			once.Do(func() {
				testApp = deployApp("example-go")
			})
		})

		It("can set and list environment variables", func() {
			sess, err := start("deis config:set POWERED_BY=midi-chlorians")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`POWERED_BY\s+midi-chlorians`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`POWERED_BY\s+midi-chlorians`))
			Eventually(sess).Should(Exit(0))

			// verify "Powered by midi-chlorians" with curl
			sess, err = start(`curl -sL "%s"; echo`, testApp.URL)
			Eventually(sess).Should(Say("Powered by midi-chlorians"))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis run env -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Say("POWERED_BY=midi-chlorians"))
			Eventually(sess).Should(Exit(0))
		})

		It("can set an integer environment variable", func() {
			sess, err := start("deis config:set FOO=1 -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`FOO\s+1`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`FOO\s+1`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis run env -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Say("FOO=1"))
			Eventually(sess).Should(Exit(0))
		})

		It("can set multiple environment variables at once", func() {
			sess, err := start("deis config:set FOO=null BAR=nil -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Exit(0))
			output := string(sess.Out.Contents())
			Expect(output).To(MatchRegexp(`FOO\s+null`))
			Expect(output).To(MatchRegexp(`BAR\s+nil`))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Exit(0))
			output = string(sess.Out.Contents())
			Expect(output).To(MatchRegexp(`FOO\s+null`))
			Expect(output).To(MatchRegexp(`BAR\s+nil`))

			sess, err = start("deis run env -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Exit(0))
			output = string(sess.Out.Contents())
			Expect(output).To(ContainSubstring("FOO=null"))
			Expect(output).To(ContainSubstring("BAR=nil"))
		})

		It("can set an environment variable containing spaces", func() {
			sess, err := start(`deis config:set -a %s POWERED_BY=the\ Deis\ team`, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`POWERED_BY\s+the Deis team`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`POWERED_BY\s+the Deis team`))
			Eventually(sess).Should(Exit(0))

			// verify "Powered by the Deis team" with curl
			sess, err = start(`curl -sL "%s"; echo`, testApp.URL)
			Eventually(sess).Should(Say("Powered by the Deis team"))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis run -a %s env", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Say("POWERED_BY=the Deis team"))
			Eventually(sess).Should(Exit(0))
		})

		It("can set a multi-line environment variable", func() {
			value := `This is
a
multiline string.`

			sess, err := start(`deis config:set -a %s FOO='%s'`, testApp.Name, value)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`FOO\s+%s`, value))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`FOO\s+%s`, value))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis run -a %s env", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Say("FOO=%s", value))
			Eventually(sess).Should(Exit(0))
		})

		It("can set an environment variable with non-ASCII and multibyte chars", func() {
			sess, err := start("deis config:set FOO=讲台 BAR=Þorbjörnsson BAZ=ноль -a %s",
				testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Exit(0))
			output := string(sess.Out.Contents())
			Expect(output).To(MatchRegexp(`FOO\s+讲台`))
			Expect(output).To(MatchRegexp(`BAR\s+Þorbjörnsson`))
			Expect(output).To(MatchRegexp(`BAZ\s+ноль`))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Exit(0))
			output = string(sess.Out.Contents())
			Expect(output).To(MatchRegexp(`FOO\s+讲台`))
			Expect(output).To(MatchRegexp(`BAR\s+Þorbjörnsson`))
			Expect(output).To(MatchRegexp(`BAZ\s+ноль`))

			sess, err = start("deis run -a %s env", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Exit(0))
			output = string(sess.Out.Contents())
			Expect(output).To(ContainSubstring("FOO=讲台"))
			Expect(output).To(ContainSubstring("BAR=Þorbjörnsson"))
			Expect(output).To(ContainSubstring("BAZ=ноль"))
		})

		It("can unset an environment variable", func() {
			sess, err := start("deis config:set -a %s FOO=xyzzy", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`FOO\s+xyzzy`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`FOO\s+xyzzy`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:unset -a %s FOO", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Removing config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Exit(0))
			Eventually(sess).ShouldNot(Say(`FOO\s+xyzzy`))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Exit(0))
			Eventually(sess).ShouldNot(Say(`FOO\s+xyzzy`))

			sess, err = start("deis run -a %s env", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Exit(0))
			Eventually(sess).ShouldNot(Say("FOO=xyzzy"))
		})

		It("can pull the configuration to an .env file", func() {
			sess, err := start("deis config:set -a %s BAZ=Freck", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Creating config"))
			Eventually(sess, defaultMaxTimeout).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`BAZ\s+Freck`))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis config:pull -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			// TODO: ginkgo seems to redirect deis' file output here, so just examine
			// the output stream rather than reading in the .env file. Bug?
			Eventually(sess).Should(Say("BAZ=Freck"))
			Eventually(sess).Should(Exit(0))
		})

		It("can push the configuration from an .env file", func() {
			contents := []byte(`BIP=baz
FOO=bar`)
			err := ioutil.WriteFile(".env", contents, 0644)
			Expect(err).NotTo(HaveOccurred())

			sess, err := start("deis config:push -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, defaultMaxTimeout).Should(Exit(0))

			sess, err = start("deis config:list -a %s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Config", testApp.Name))
			Eventually(sess).Should(Say(`BIP\s+baz`))
			Eventually(sess).Should(Exit(0))
		})
	})

	DescribeTable("can get command-line help for config", func(cmd, expected string) {

		sess, err := start(cmd)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Say(expected))
		Eventually(sess).Should(Exit(0))
		// TODO: test that help output was more than five lines long
	},

		Entry("helps on \"help config\"",
			"deis help config", "Valid commands for config:"),
		Entry("helps on \"config -h\"",
			"deis config -h", "Valid commands for config:"),
		Entry("helps on \"config --help\"",
			"deis config --help", "Valid commands for config:"),
		Entry("helps on \"help config:list\"",
			"deis help config:list", "Lists environment variables for an application."),
		Entry("helps on \"config:list -h\"",
			"deis config:list -h", "Lists environment variables for an application."),
		Entry("helps on \"config:list --help\"",
			"deis config:list --help", "Lists environment variables for an application."),
		Entry("helps on \"help config:set\"",
			"deis help config:set", "Sets environment variables for an application."),
		Entry("helps on \"config:set -h\"",
			"deis config:set -h", "Sets environment variables for an application."),
		Entry("helps on \"config:set --help\"",
			"deis config:set --help", "Sets environment variables for an application."),
		Entry("helps on \"help config:unset\"",
			"deis help config:unset", "Unsets an environment variable for an application."),
		Entry("helps on \"config:unset -h\"",
			"deis config:unset -h", "Unsets an environment variable for an application."),
		Entry("helps on \"config:unset --help\"",
			"deis config:unset --help", "Unsets an environment variable for an application."),
		Entry("helps on \"help config:pull\"",
			"deis help config:pull", "Extract all environment variables from an application for local use."),
		Entry("helps on \"config:pull -h\"",
			"deis config:pull -h", "Extract all environment variables from an application for local use."),
		Entry("helps on \"config:pull --help\"",
			"deis config:pull --help", "Extract all environment variables from an application for local use."),
		Entry("helps on \"help config:push\"",
			"deis help config:push", "Sets environment variables for an application."),
		Entry("helps on \"config:push -h\"",
			"deis config:push -h", "Sets environment variables for an application."),
		Entry("helps on \"config:push --help\"",
			"deis config:push --help", "Sets environment variables for an application."),
	)
})
