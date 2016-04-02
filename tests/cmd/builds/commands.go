package builds

import (
	"time"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis builds` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

const ExampleImage = "deis/example-dockerfile-http"

// Create executes `deis builds:create` as the specified user.
func Create(user model.User, app model.App) {
	createOrPull(user, app, "builds:create")
}

// Pull executes the `deis pull` shortcut as the specified user.
func Pull(user model.User, app model.App) {
	createOrPull(user, app, "pull")
}

func createOrPull(user model.User, app model.App, command string) {
	sess, err := cmd.Start("deis %s --app=%s %s", &user, command, app.Name, ExampleImage)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("Creating build..."))
	Eventually(sess).Should(Exit(0))
	time.Sleep(10 * time.Second)
}
