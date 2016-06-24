package tests

import (
	"regexp"
	"strings"

	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis tags", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app that has already been deployed", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
				builds.Create(user, app)
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list that app's tags", func() {
				sess, err := cmd.Start("deis tags:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Tags", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user cannot set an invalid tag", func() {
				sess, err := cmd.Start("deis tags:set --app=%s munkafolyamat=yeah", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).ShouldNot(Say("=== %s Tags", app.Name))
				Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, app.Name))
				Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrTagNotFound)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user cannot unset an invalid tag", func() {
				sess, err := cmd.Start("deis tags:unset --app=%s munkafolyamat", &user, app.Name)
				Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user can set a valid tag", func() {
				// Find a valid tag to set
				// Use original $HOME dir or else kubectl can't find its config
				sess, err := cmd.Start("HOME=%s kubectl get nodes -o jsonpath={.items[*].metadata..labels}", nil, settings.ActualHome)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// grep output like "map[kubernetes.io/hostname:192.168.64.2 node:worker1]"
				re := regexp.MustCompile(`([\w\.\-]{0,253}/?[-_\.\w]{1,63}:[-_\.\w]{1,63})`)
				pairs := re.FindAllString(string(sess.Out.Contents()), -1)
				// Use the first key:value pair found
				label := strings.Split(pairs[0], ":")

				sess, err = cmd.Start("deis tags:set --app=%s %s=%s", &user, app.Name, label[0], label[1])
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Tags", app.Name))
				Eventually(sess).Should(Say(`%s\s+%s`, label[0], label[1]))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis tags:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Tags", app.Name))
				Eventually(sess).Should(Say(`%s\s+%s`, label[0], label[1]))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and a tag has already been added to the app", func() {

				var label []string

				BeforeEach(func() {
					// Find a valid tag to set
					// Use original $HOME dir or else kubectl can't find its config
					sess, err := cmd.Start("HOME=%s kubectl get nodes -o jsonpath={.items[*].metadata..labels}", nil, settings.ActualHome)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					// grep output like "map[kubernetes.io/hostname:192.168.64.2 node:worker1]"
					re := regexp.MustCompile(`([\w\.\-]{0,253}/?[-_\.\w]{1,63}:[-_\.\w]{1,63})`)
					pairs := re.FindAllString(string(sess.Out.Contents()), -1)
					// Use the first key:value pair found
					label = strings.Split(pairs[0], ":")

					sess, err = cmd.Start("deis tags:set --app=%s %s=%s", &user, app.Name, label[0], label[1])
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Tags", app.Name))
					Eventually(sess).Should(Say(`%s\s+%s`, label[0], label[1]))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can unset that tag from that app", func() {
					sess, err := cmd.Start("deis tags:unset --app=%s %s", &user, app.Name, label[0])
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Tags", app.Name))
					Eventually(sess).ShouldNot(Say(`%s\s+%s`, label[0], label[1]))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis tags:list --app=%s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Tags", app.Name))
					Eventually(sess).ShouldNot(Say(`%s\s+%s`, label[0], label[1]))
					Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

	})

	DescribeTable("any user can get command-line help for tags", func(command string, expected string) {
		sess, err := cmd.Start(command, nil)
		Eventually(sess).Should(Say(expected))
		Expect(err).NotTo(HaveOccurred())
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
