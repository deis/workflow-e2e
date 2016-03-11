package tests

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Tags", func() {

	Context("with a deployed app", func() {
		var testApp App
		once := &sync.Once{}

		BeforeEach(func() {
			// use the "kubectl" executable in the search $PATH
			if _, err := exec.LookPath("kubectl"); err != nil {
				Skip("kubectl not found in search $PATH")
			}

			// Set up the Tags test app only once and assume the suite will clean up.
			once.Do(func() {
				os.Chdir("example-go")
				appName := getRandAppName()
				createApp(appName)
				testApp = deployApp(appName)
			})
		})

		It("can set and unset tags", func() {
			// can list tags
			sess, err := start("deis tags:list")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).Should(Exit(0))

			// set an invalid tag
			sess, err = start("deis tags:set munkafolyamat=yeah")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).ShouldNot(Say("=== %s Tags", testApp.Name))
			Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, testApp.Name))
			Eventually(sess.Err).Should(Say("400 Bad Request"))
			Eventually(sess).Should(Exit(1))

			// list tags
			sess, err = start("deis tags:list")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, testApp.Name))
			Eventually(sess).Should(Exit(0))

			// find a valid tag to set
			cmd := "kubectl get nodes -o jsonpath={.items[*].metadata..labels}"
			// use original $HOME dir or kubectl can't find its config
			sess, err = start("HOME=%s %s", homeHome, cmd)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			// grep output like "map[kubernetes.io/hostname:192.168.64.2 node:worker1]"
			re := regexp.MustCompile(`([\w\.]{0,253}/?[-_\.\w]{1,63}:[-_\.\w]{1,63})`)
			pairs := re.FindAllString(string(sess.Out.Contents()), -1)
			// use the first key:value pair found
			label := strings.Split(pairs[0], ":")

			// set a valid tag
			sess, err = start("deis tags:set %s=%s", label[0], label[1])
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, "5m").Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).Should(Say(`%s\s+%s`, label[0], label[1]))
			Eventually(sess).Should(Exit(0))

			// list tags
			sess, err = start("deis tags:list")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).Should(Say(`%s\s+%s`, label[0], label[1]))
			Eventually(sess).Should(Exit(0))

			// unset an invalid tag
			sess, err = start("deis tags:unset munkafolyamat")
			Expect(err).NotTo(HaveOccurred())
			// TODO: should unsetting a bogus tag return 0 (success?)
			Eventually(sess, "5m").Should(Exit(0))
			Eventually(sess).Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, testApp.Name))

			// unset a valid tag
			sess, err = start("deis tags:unset %s", label[0])
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, "5m").Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).Should(Exit(0))
			Eventually(sess).ShouldNot(Say(`%s\s+%s`, label[0], label[1]))

			// list tags
			sess, err = start("deis tags:list")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Tags", testApp.Name))
			Eventually(sess).Should(Exit(0))
			Eventually(sess).ShouldNot(Say(`%s\s+%s`, label[0], label[1]))
			Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, testApp.Name))
		})
	})

	DescribeTable("can get command-line help for tags", func(cmd, expected string) {

		sess, err := start(cmd)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Say(expected))
		Eventually(sess).Should(Exit(0))
		// TODO: test that help output was more than five lines long
	},

		Entry("helps on \"help tags\"",
			"deis help tags", "Valid commands for tags:"),
		Entry("helps on \"tags -h\"",
			"deis tags -h", "Valid commands for tags:"),
		Entry("helps on \"tags --help\"",
			"deis tags --help", "Valid commands for tags:"),
		Entry("helps on \"help tags:list\"",
			"deis help tags:list", "Lists tags for an application."),
		Entry("helps on \"tags:list -h\"",
			"deis tags:list -h", "Lists tags for an application."),
		Entry("helps on \"tags:list --help\"",
			"deis tags:list --help", "Lists tags for an application."),
		Entry("helps on \"help tags:set\"",
			"deis help tags:set", "Sets tags for an application."),
		Entry("helps on \"tags:set -h\"",
			"deis tags:set -h", "Sets tags for an application."),
		Entry("helps on \"tags:set --help\"",
			"deis tags:set --help", "Sets tags for an application."),
		Entry("helps on \"help tags:unset\"",
			"deis help tags:unset", "Unsets tags for an application."),
		Entry("helps on \"tags:unset -h\"",
			"deis tags:unset -h", "Unsets tags for an application."),
		Entry("helps on \"tags:unset --help\"",
			"deis tags:unset --help", "Unsets tags for an application."),
	)

})
