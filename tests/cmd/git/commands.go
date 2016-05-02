package git

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `git` commands.
// This allows each of these to be re-used easily in multiple contexts.

// Push executes a `git push deis master` from the current directory using the provided key.
func Push(user model.User, keyPath string, app model.App, banner string) {
	sess, err := cmd.Start("GIT_SSH=%s GIT_KEY=%s git push deis master", &user, settings.GitSSH, keyPath)
	Expect(err).NotTo(HaveOccurred())
	// sess.Wait(settings.MaxEventuallyTimeout)
	// output := string(sess.Out.Contents())
	// Expect(output).To(MatchRegexp(`Done, %s:v\d deployed to Deis`, app.Name))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
	// curl the app's root URL and print just the HTTP response code
	cmdRetryTimeout := 60
	curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(
		`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
	Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
	// verify that the response contains "Powered by" as all the example apps do
	curlCmd = model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL "%s"`, app.URL)}
	Eventually(cmd.Retry(curlCmd, banner, cmdRetryTimeout)).Should(BeTrue())
}
