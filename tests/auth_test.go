package tests

import (
	"fmt"
	"os"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	gexpect "github.com/ThomasRooney/gexpect"
)

var _ = Describe("deis auth", func() {

	Context("with no user logged in", func() {

		BeforeEach(func() {
			// Important: All the tests use profiles. In theory, no client.json containing a token
			// exists because of this. However, in order to future-proof this test against any fallout
			// from any test added in the future that might deliberately or accidentally behave
			// differently, we explicitly log out, without specifying a profile. This is meant to
			// GUARANTEE that client.json does not exist.
			sess, err := cmd.Start("deis auth:logout", nil)
			Eventually(sess).Should(Say("Logged out\n"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Specify("information on the current user cannot be printed", func() {
			sess, err := cmd.Start("deis auth:whoami", nil)
			Eventually(sess.Err).Should(Say("Error: Client configuration file not found"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})

	})

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = model.NewUser()
			os.Setenv("DEIS_PROFILE", user.Username)
		})

		AfterEach(func() {
			auth.Cancel(user)
			os.Unsetenv("DEIS_PROFILE")
		})

		Specify("that user can register in an interactive manner", func() {
			sess, err := gexpect.Spawn(fmt.Sprintf("deis auth:register %s --password=%s", settings.DeisControllerURL, user.Password))
			Expect(err).NotTo(HaveOccurred())

			err = sess.Expect("username:")
			Expect(err).NotTo(HaveOccurred())
			sess.SendLine(user.Username)

			err = sess.Expect("email:")
			Expect(err).NotTo(HaveOccurred())
			sess.SendLine(user.Email)

			sess.Expect(fmt.Sprintf("Registered %s", user.Username))
			Expect(err).NotTo(HaveOccurred())

			sess.Expect(fmt.Sprintf("Logged in as %s", user.Username))
			Expect(err).NotTo(HaveOccurred())

			auth.Whoami(user)
		})
	})

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Specify("that user can log out", func() {
			auth.Logout(user)
			auth.Login(user) // Log back in so cleanup won't fail.
		})

		Specify("a new user cannot register using the same details", func() {
			sess, err := cmd.Start("deis auth:register %s --username=%s --password=%s --email=%s", nil, settings.DeisControllerURL, user.Username, user.Password, user.Email)
			Eventually(sess.Err).Should(Say("Registration failed"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})

		Specify("that user can print information about themself", func() {
			auth.Whoami(user)
		})

		Specify("that user can print extensive information about themself", func() {
			auth.WhoamiAll(user)
		})

		Specify("that user can regenerates their own token", func() {
			auth.Regenerate(user)
		})

	})

	Context("with an existing admin", func() {

		admin := model.Admin

		Specify("that admin can list admins", func() {
			sess, err := cmd.Start("deis perms:list --admin", &admin)
			Eventually(sess).Should(Say("=== Administrators"))
			Eventually(sess).Should(Say(admin.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Context("and another existing user", func() {

			var otherUser model.User

			BeforeEach(func() {
				otherUser = auth.Register()
			})

			AfterEach(func() {
				auth.Cancel(otherUser)
			})

			Specify("that admin can regenerate the token for the other user", func() {
				sess, err := cmd.Start("deis auth:regenerate -u %s", &admin, otherUser.Username)
				Eventually(sess).Should(Say("Token Regenerated"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
				auth.Login(otherUser) // Log back in so cleanup won't fail.
			})

		})

		// TODO: This is marked pending because it resets all user auth tokens. Because we run the
		// tests in parallel, this can wreak havoc on tests that may be in flight. We will need to
		// reevaluate how we want to test this functionality.
		XSpecify("that admin can regenerate the tokens of all other users", func() {
			sess, err := cmd.Start("deis auth:regenerate --all", &admin)
			Eventually(sess).Should(Say("Token Regenerated"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

	})

})
