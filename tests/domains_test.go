package tests

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/cmd/domains"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis domains", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list that app's domains", func() {
				sess, err := cmd.Start("deis domains:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Domains", app.Name))
				Eventually(sess).Should(Say("%s", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can add domains to that app", func() {
				domain := getRandDomain()
				sess, err := cmd.Start("deis domains:add %s --app=%s", &user, domain, app.Name)
				Eventually(sess).Should(Say("Adding %s to %s...", domain, app.Name))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user cannot remove a non-existent domain from that app", func() {
				sess, err := cmd.Start("deis domains:remove --app=%s %s", &user, app.Name, "non.existent.domain")
				Eventually(sess.Err, settings.MaxEventuallyTimeout).Should(Say(util.PrependError(domains.ErrNoDomainMatch)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Context("with a domain added to it", func() {

				var domain string

				BeforeEach(func() {
					domain = getRandDomain()
					domains.Add(user, app, domain)
				})

				Specify("that user can remove that domain from that app", func() {
					sess, err := cmd.Start("deis domains:remove %s --app=%s", &user, domain, app.Name)
					Eventually(sess).Should(Say("Removing %s from %s...", domain, app.Name))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

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

			Context("with a domain added to it", func() {

				cmdRetryTimeout := 60

				var domain string

				BeforeEach(func() {
					domain = getRandDomain()
					domains.Add(user, app, domain)
				})

				AfterEach(func() {
					domains.Remove(user, app, domain)
					// App can no longer be accessed at the previously associated domain
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -H "Host: %s" -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusNotFound), cmdRetryTimeout)).Should(BeTrue())
				})

				Specify("that app can be accessed at its usual address", func() {
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
				})

				Specify("that app can be accessed at the associated domain", func() {
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -H "Host: %s" -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
				})

			})

		})

	})

})

func getRandDomain() string {
	return fmt.Sprintf("my-custom-%d.domain.com", rand.Intn(999999999))
}
