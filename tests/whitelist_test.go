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

var _ = Describe("deis whitelist", func() {

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

			Specify("can list that app's whitelist list", func() {
				sess, err := cmd.Start("deis whitelist:list --app=%s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("can view app when no addresses whitelist", func() {
				// curl the app's root URL and print just the HTTP response code
				cmdRetryTimeout := 60
				curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(200), cmdRetryTimeout)).Should(BeTrue())
			})

			Specify("can add/remove addresses from the whitelist", func() {
				sess, err := cmd.Start("deis whitelist:add 1.2.3.4 --app=%s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Eventually(sess).Should(Exit(0))

				// curl the app's root URL and print just the HTTP response code
				cmdRetryTimeout := 60
				curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(403), cmdRetryTimeout)).Should(BeTrue())

				sess, err = cmd.Start("deis whitelist:add 0.0.0.0/0 --app=%s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Eventually(sess).Should(Exit(0))

				cmdRetryTimeout = 60
				curlCmd = model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(200), cmdRetryTimeout)).Should(BeTrue())

				sess, err = cmd.Start("deis whitelist:remove 0.0.0.0/0 --app=%s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Eventually(sess).Should(Exit(0))

				// curl the app's root URL and print just the HTTP response code
				cmdRetryTimeout = 60
				curlCmd = model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
				Eventually(cmd.Retry(curlCmd, strconv.Itoa(403), cmdRetryTimeout)).Should(BeTrue())
			})
		})
	})

})
