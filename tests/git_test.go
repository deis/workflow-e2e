package tests

import (
	"os"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis git", func() {

	Context("with an existing user", func() {

		var user model.User
		const dir = "git-test"

		BeforeEach(func() {
			user = auth.Register()

			err := os.Mkdir(dir, os.ModeDir)
			Expect(err).NotTo(HaveOccurred())

			err = os.Chdir(dir)
			Expect(err).NotTo(HaveOccurred())

			output, err := cmd.Execute("git init")
			Expect(err).NotTo(HaveOccurred(), output)
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Specify("will not face an error when trying to add an existing correct remote", func() {
			app := apps.Create(user)
			defer apps.Destroy(user, app)
			sess, err := cmd.Start("deis git:remote", &user)
			Eventually(sess).Should(Say("Remote deis already exists and is correctly configured for app"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Specify("can destroy and recreate a remote", func() {
			app := apps.Create(user)
			defer apps.Destroy(user, app)
			sess, err := cmd.Start("deis git:remove", &user)
			Eventually(sess).Should(Say("Git remotes for app %s removed.", app.Name))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
			sess, err = cmd.Start("deis git:remote", &user)
			Eventually(sess).Should(Say("Git remote deis successfully created for app"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Specify("will face an error when overwriting a remote for an existing app but can force", func() {
			app := apps.Create(user)
			defer apps.Destroy(user, app)
			sess, err := cmd.Start("deis git:remote --app=foo", &user)
			Eventually(sess).Should(Say("Error: Remote deis already exists, please run"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
			sess, err = cmd.Start("deis git:remote -f --app=foo", &user)
			Eventually(sess).Should(Say("Git remote deis successfully created for app foo."))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})
	})
})
