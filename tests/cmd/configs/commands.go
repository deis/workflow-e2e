package configs

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis config` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// Set executes `deis config:set` on the specified app as the specified user.
func Set(user model.User, app model.App, key string, value string) *Session {
	sess, err := cmd.Start("deis config:set %s=%s --app=%s", &user, key, value, app.Name)
	Expect(err).NotTo(HaveOccurred())
	sess.Wait(settings.MaxEventuallyTimeout)
	Eventually(sess).Should(Say("Creating config..."))
	Eventually(sess).Should(Exit(0))
	return sess
}
