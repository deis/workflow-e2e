package tests

import (
	"os"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
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

	Context("with a non-admin user", func() {

		var user model.User

		BeforeEach(func() {
			user = model.NewUser()
			os.Setenv("DEIS_PROFILE", user.Username)
		})

		AfterEach(func() {
			sess, err := cmd.Start("deis auth:cancel --username=%s --password=%s --yes", &user, user.Username, user.Password)
			Expect(err).To(BeNil())
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
			os.Unsetenv("DEIS_PROFILE")
		})

		Specify("that user cannot register when registration mode is 'admin_only', as is the default", func() {
			sess, err := cmd.Start("deis auth:register %s --username=%s --password=%s --email=%s", nil, settings.DeisControllerURL, user.Username, user.Password, user.Email)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("Registration failed: Error: You do not have permission to perform this action."))
			Eventually(sess).Should(Exit(1))
		})

	})

	Context("with an existing user", func() {
		admin := model.Admin
		var user model.User

		BeforeEach(func() {
			user = auth.RegisterAndLogin()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Specify("that user can log out", func() {
			auth.Logout(user)
			auth.Login(user) // Log back in so cleanup won't fail.
		})

		Specify("a new user cannot be registered using the same details", func() {
			sess, err := cmd.Start("deis auth:register %s --username=%s --password=%s --email=%s", &admin, settings.DeisControllerURL, user.Username, user.Password, user.Email)
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
				otherUser = auth.RegisterAndLogin()
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
