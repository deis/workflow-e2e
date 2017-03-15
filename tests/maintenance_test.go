package tests

import (
	"fmt"
	"strconv"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis maintenance", func() {

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

			Specify("can list that app's maintenance info", func() {
				sess, err := cmd.Start("deis maintenance:info --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("Maintenance mode is off.\n"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("can view app when maintenance mode is off", func() {
				// curl the app's root URL and print just the HTTP response code
				cmdRetryTimeout := 60
				curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(200), cmdRetryTimeout)).Should(BeTrue())
			})

			Specify("can enable/disable maintenance", func() {
				sess, err := cmd.Start("deis maintenance:on --app=%s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis maintenance:info --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("Maintenance mode is on.\n"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// curl the app's root URL and print just the HTTP response code
				cmdRetryTimeout := 60
				curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(503), cmdRetryTimeout)).Should(BeTrue())

				sess, err = cmd.Start("deis maintenance:off --app=%s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis maintenance:info --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("Maintenance mode is off.\n"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				cmdRetryTimeout = 60
				curlCmd = model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(200), cmdRetryTimeout)).Should(BeTrue())
			})
		})
	})

})
