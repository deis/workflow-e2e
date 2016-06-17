package apps

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/deis/workflow-e2e/shims"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis apps` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// Create executes `deis apps:create` as the specified user with the specified, arvitrary options.
func Create(user model.User, options ...string) model.App {
	noRemote := false
	app := model.NewApp()
	sess, err := cmd.Start("deis apps:create %s %s", &user, app.Name, strings.Join(options, " "))
	Expect(err).NotTo(HaveOccurred())
	sess.Wait(settings.MaxEventuallyTimeout)
	Eventually(sess).Should(Say("created %s", app.Name))

	for _, option := range options {
		if option == "--no-remote" {
			noRemote = true
			break
		}
	}

	if noRemote {
		Eventually(sess).Should(Say("If you want to add a git remote for this app later, use "))
	} else {
		Eventually(sess).Should(Say("Git remote deis added"))
		Eventually(sess).Should(Say("remote available at "))
	}
	Eventually(sess).Should(Exit(0))
	return app
}

// Open executes `deis apps:open` on the specified app as the specified user. A shim is used to
// intercept the execution of `open` (Darwin) or `xdg-open` (Linux) and verify that the browser
// would have navigated to the correct address.
func Open(user model.User, app model.App) {
	// The underlying utility that `deis open` looks for:
	toShim := "open" //darwin
	if runtime.GOOS == "linux" {
		toShim = "xdg-open"
	}
	myShim, err := shims.CreateSystemShim(toShim)
	if err != nil {
		panic(err)
	}
	defer shims.RemoveShim(myShim)

	// Create custom env with location of open shim prepended to the PATH env var.
	env := shims.PrependPath(os.Environ(), os.TempDir())

	sess, err := cmd.StartCmd(model.Cmd{Env: env, CommandLineString: fmt.Sprintf("DEIS_PROFILE=%s deis open -a %s", user.Username, app.Name)})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))

	output, err := ioutil.ReadFile(myShim.OutFile.Name())
	Expect(err).NotTo(HaveOccurred())
	Expect(strings.TrimSpace(string(output))).To(ContainSubstring(app.URL))
}

// Destroy executes `deis apps:destroy` on the specified app as the specified user.
func Destroy(user model.User, app model.App) *Session {
	sess, err := cmd.Start("deis apps:destroy --app=%s --confirm=%s", &user, app.Name, app.Name)
	Expect(err).NotTo(HaveOccurred())
	sess.Wait(settings.MaxEventuallyTimeout)
	Eventually(sess).Should(Say("Destroying %s...", app.Name))
	Eventually(sess).Should(Say(`done in `))
	Eventually(sess).Should(Exit(0))
	return sess
}
