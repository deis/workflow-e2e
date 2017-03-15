package tests

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis routing", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.RegisterAndLogin()
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

			Specify("that user can list that app's routing info", func() {
				sess, err := cmd.Start("deis routing:info --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("Routing is enabled.\n"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can view app when routing is enabled", func() {
				cmdRetryTimeout := 60
				curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
			})

			Specify("that user can disable routing", func() {
				cmdRetryTimeout := 60
				sess, err := cmd.Start("deis routing:disable --app=%s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis routing:info --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("Routing is disabled.\n"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusNotFound), cmdRetryTimeout)).Should(BeTrue())
			})
		})
	})

	DescribeTable("any user can get command-line help for routing", func(command string, expected string) {
		sess, err := cmd.Start(command, nil)
		Eventually(sess).Should(Say(expected))
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Exit(0))
		// TODO: test that help output was more than five lines long
	},
		Entry("helps on \"help routing\"",
			"deis help routing", "Valid commands for routing:"),
		Entry("helps on \"routing -h\"",
			"deis routing -h", "Valid commands for routing:"),
		Entry("helps on \"routing --help\"",
			"deis routing --help", "Valid commands for routing:"),
		Entry("helps on \"help routing:info\"",
			"deis help routing:info", "Prints info about the current application's routability."),
		Entry("helps on \"routing:info -h\"",
			"deis routing:info -h", "Prints info about the current application's routability."),
		Entry("helps on \"routing:info --help\"",
			"deis routing:info --help", "Prints info about the current application's routability."),
		Entry("helps on \"help routing:enable\"",
			"deis help routing:enable", "Enables routability for an app."),
		Entry("helps on \"routing:enable -h\"",
			"deis routing:enable -h", "Enables routability for an app."),
		Entry("helps on \"routing:enable --help\"",
			"deis routing:enable --help", "Enables routability for an app."),
		Entry("helps on \"help routing:disable\"",
			"deis help routing:disable", "Disables routability for an app."),
		Entry("helps on \"routing:disable -h\"",
			"deis routing:disable -h", "Disables routability for an app."),
		Entry("helps on \"routing:disable --help\"",
			"deis routing:disable --help", "Disables routability for an app."),
	)

})
