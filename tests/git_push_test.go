package tests

import (
	"os"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/git"
	"github.com/deis/workflow-e2e/tests/cmd/keys"
	"github.com/deis/workflow-e2e/tests/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("git push deis master", func() {

	Context("with an existing user", func() {

		var user model.User
		var keyPath string

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who has added their public key", func() {

			BeforeEach(func() {
				_, keyPath = keys.Add(user)
			})

			Context("and who has a local git repo containing source code", func() {

				BeforeEach(func() {
					output, err := cmd.Execute(`git clone https://github.com/deis/example-go.git`)
					Expect(err).NotTo(HaveOccurred(), output)
				})

				Context("and has run `deis apps:create` from within that repo", func() {

					var app model.App

					BeforeEach(func() {
						os.Chdir("example-go")
						app = apps.Create(user)
					})

					AfterEach(func() {
						apps.Destroy(user, app)
					})

					Specify("that user can deploy that app using a git push", func() {
						git.Push(user, keyPath, app)
					})

				})

			})

		})

	})

})
