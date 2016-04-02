package perms

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis perms` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// Create executes `deis perms:create` as the specified user to grant permissions on the specified
// app to a second user.
func Create(user model.User, app model.App, grantUser model.User) {
	sess, err := cmd.Start("deis perms:create %s --app=%s", &user, grantUser.Username, app.Name)
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to %s collaborators... done\n", grantUser.Username, app.Name))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
}

// Delete executes `deis perms:delete` as the specified user to revoke permissions on the specified
// app from a second user.
func Delete(user model.User, app model.App, revokeUser model.User) {
	sess, err := cmd.Start("deis perms:delete %s --app=%s", &user, revokeUser.Username, app.Name)
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Removing %s from %s collaborators... done", revokeUser.Username, app.Name))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
}
