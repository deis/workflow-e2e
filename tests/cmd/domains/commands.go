package domains

import (
	"fmt"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis domains` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// Add executes `deis domains:add` as the specified user to add the specified domain to the
// specified app.
func Add(user model.User, app model.App, domain string) {
	sess, err := cmd.Start("deis domains:add %s --app=%s", &user, domain, app.Name)
	// Explicitly build literal substring since 'domain' may be a wildcard domain ('*.foo.com') and
	// we don't want Gomega interpreting this string as a regexp
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(fmt.Sprintf("Adding %s to %s...", domain, app.Name)))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
}

// Remove executes `deis domains:remove` as the specified user to remove the specified domain from
// the specified app.
func Remove(user model.User, app model.App, domain string) {
	sess, err := cmd.Start("deis domains:remove %s --app=%s", &user, domain, app.Name)
	// Explicitly build literal substring since 'domain' may be a wildcard domain ('*.foo.com') and
	// we don't want Gomega interpreting this string as a regexp
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(fmt.Sprintf("Removing %s from %s...", domain, app.Name)))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
}
