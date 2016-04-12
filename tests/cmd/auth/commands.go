package auth

import (
	"fmt"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis auth` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// RegisterAdmin executes `deis auth:register`, using hard-coded username, password, and email
// address. When this is executed, it is executed in hopes of registering Workflow's FIRST user,
// which will automatically have admin permissions. If this should fail, the function proceeds
// with logging in using those same hard-coded credentials, in the hopes that the reason for the
// failure is that such an account already exists, having been created by a previous execution of
// the tests.
func RegisterAdmin() {
	admin := model.Admin
	sess, err := cmd.Start("deis auth:register %s --username=%s --password=%s --email=%s", &admin, settings.DeisControllerURL, admin.Username, admin.Password, admin.Email)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit())
	Expect(err).NotTo(HaveOccurred())

	// We cannot entirely count on the registration having succeeded. It may have failed if a user
	// with the username "admin" already exists. However, if that user IS indeed an admin and their
	// password is also "admin" (e.g. the admin was created by a previous run of these tests), then
	// we can proceed... so attempt to login...
	Login(admin)

	// Now verify this user is an admin by running a privileged command.
	sess, err = cmd.Start("deis users:list", &admin)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

// Register executes `deis auth:register` using a randomized username and returns a model.User.
func Register() model.User {
	user := model.NewUser()
	sess, err := cmd.Start("deis auth:register %s --username=%s --password=%s --email=%s", &user, settings.DeisControllerURL, user.Username, user.Password, user.Email)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say(fmt.Sprintf("Logged in as %s\n", user.Username)))
	return user
}

// Login executes `deis auth:login` as the specified user. In the process, it creates the a
// corresponding profile that contains the user's authentication token. Re-use of this profile is
// for most other actions is what permits multiple test users to act in parallel without impacting
// one another.
func Login(user model.User) {
	sess, err := cmd.Start("deis auth:login %s --username=%s --password=%s", &user, settings.DeisControllerURL, user.Username, user.Password)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say(fmt.Sprintf("Logged in as %s\n", user.Username)))
}

// Whoami executes `deis auth:whoami` as the specified user.
func Whoami(user model.User) {
	sess, err := cmd.Start("deis auth:whoami", &user)
	Eventually(sess).Should(Say("You are %s", user.Username))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

// Regenerate executes `deis auth:regenerate` as the specified user.
func Regenerate(user model.User) {
	sess, err := cmd.Start("deis auth:regenerate", &user)
	Eventually(sess).Should(Say("Token Regenerated"))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

// Logout executes `deis auth:logout` as the specified user.
func Logout(user model.User) {
	sess, err := cmd.Start("deis auth:logout", &user)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("Logged out\n"))
}

// Cancel executes `deis auth:cancel` as the specified user.
func Cancel(user model.User) {
	sess, err := cmd.Start("deis auth:cancel --username=%s --password=%s --yes", &user, user.Username, user.Password)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("Account cancelled\n"))
}

// CancelAdmin deletes the admin user that was created to facilitate the tests.
func CancelAdmin() {
	admin := model.Admin
	sess, err := cmd.Start("deis auth:cancel --username=%s --password=%s --yes", &admin, admin.Username, admin.Password)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("Account cancelled\n"))
}
