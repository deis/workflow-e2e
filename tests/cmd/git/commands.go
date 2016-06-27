package git

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `git` commands.
// This allows each of these to be re-used easily in multiple contexts.

const (
	pushCommandLineString = "GIT_SSH=%s GIT_KEY=%s git push deis master"
)

// Push executes a `git push deis master` from the current directory using the provided key.
func Push(user model.User, keyPath string, app model.App, banner string) {
	sess := StartPush(user, keyPath)
	// sess.Wait(settings.MaxEventuallyTimeout)
	// output := string(sess.Out.Contents())
	// Expect(output).To(MatchRegexp(`Done, %s:v\d deployed to Deis`, app.Name))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
	Curl(app, banner)
}

// Curl polls an app over HTTP until it returns the expected "Powered by" banner.
func Curl(app model.App, banner string) {
	// curl the app's root URL and print just the HTTP response code
	cmdRetryTimeout := 60
	curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(
		`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
	Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
	// verify that the response contains "Powered by" as all the example apps do
	curlCmd = model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL "%s"`, app.URL)}
	Eventually(cmd.Retry(curlCmd, banner, cmdRetryTimeout)).Should(BeTrue())
}

// PushWithInterrupt executes a `git push deis master` from the current
// directory using the provided key, but then halts the progress via SIGINT.
func PushWithInterrupt(user model.User, keyPath string) {
	sess := StartPush(user, keyPath)
	Eventually(sess.Err).Should(Say("Starting build... but first, coffee!"))

	sess = sess.Interrupt()

	newSess := StartPush(user, keyPath)
	Eventually(newSess.Err).ShouldNot(Say("exec request failed on channel 0"))
	Eventually(newSess.Err).Should(Say("fatal: remote error: Another git push is ongoing"))
	Eventually(newSess, settings.DefaultEventuallyTimeout).Should(Exit(128))
}

// PushUntilResult executes a `git push deis master` from the current
// directory using the provided key, until the command result satisfies
// expectedCmdResult of type model.CmdResult, failing if
// settings.DefaultEventuallyTimeout is reached first.
func PushUntilResult(user model.User, keyPath string, expectedCmdResult model.CmdResult) {
	envVars := append(os.Environ(), fmt.Sprintf("DEIS_PROFILE=%s", user.Username))
	pushCmd := model.Cmd{Env: envVars, CommandLineString: fmt.Sprintf(
		pushCommandLineString, settings.GitSSH, keyPath)}

	Eventually(cmd.RetryUntilResult(pushCmd, expectedCmdResult, 5*time.Second,
		settings.MaxEventuallyTimeout)).Should(BeTrue())
}

// StartPush starts a `git push deis master` command and returns the command session.
func StartPush(user model.User, keyPath string) *Session {
	sess, err := cmd.Start(pushCommandLineString, &user, settings.GitSSH, keyPath)
	Expect(err).NotTo(HaveOccurred())
	return sess
}
