package git

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `git` commands.
// This allows each of these to be re-used easily in multiple contexts.

// Push executes a `git push deis master` from the current directory using the provided key.
func Push(user model.User, keyPath string) {
	sess, err := cmd.Start("GIT_SSH=%s GIT_KEY=%s git push deis master", &user, settings.GitSSH, keyPath)
	Expect(err).NotTo(HaveOccurred())
	// sess.Wait(settings.MaxEventuallyTimeout)
	// output := string(sess.Out.Contents())
	// Expect(output).To(MatchRegexp(`Done, %s:v\d deployed to Deis`, app.Name))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
}
