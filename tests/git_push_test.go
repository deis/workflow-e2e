package tests

import (
	"os"
	"path/filepath"
	"time"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/git"
	"github.com/deis/workflow-e2e/tests/cmd/keys"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
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

			Context("and who has a local git repo containing buildpack source code", func() {

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
						git.Push(user, keyPath, app, "Powered by Deis")
					})

					Specify("that user can interrupt the deploy of the app and recover", func() {
						git.PushWithInterrupt(user, keyPath)

						git.PushUntilResult(user, keyPath,
							model.CmdResult{
								Out:      nil,
								Err:      []byte("Everything up-to-date"),
								ExitCode: 0,
							})
					})

					Specify("that user can deploy that app only once concurrently", func() {
						sess := git.StartPush(user, keyPath)
						// sleep for five seconds, then push the same app
						time.Sleep(5000 * time.Millisecond)
						sess2 := git.StartPush(user, keyPath)
						Eventually(sess2.Err).Should(Say("fatal: remote error: Another git push is ongoing"))
						Eventually(sess2).Should(Exit(128))
						git.PushUntilResult(user, keyPath,
							model.CmdResult{
								Out:      nil,
								Err:      []byte("Everything up-to-date"),
								ExitCode: 0,
							})
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
						git.Curl(app, "Powered by Deis")
					})

					Context("and who has another local git repo containing buildpack source code", func() {

						BeforeEach(func() {
							os.Chdir("..")
							output, err := cmd.Execute(`git clone https://github.com/deis/example-nodejs-express.git`)
							Expect(err).NotTo(HaveOccurred(), output)
						})

						Context("and has run `deis apps:create` from within that repo", func() {

							var app2 model.App

							BeforeEach(func() {
								os.Chdir("example-nodejs-express")
								app2 = apps.Create(user)
							})

							AfterEach(func() {
								apps.Destroy(user, app2)
							})

							Specify("that user can deploy both apps concurrently", func() {
								os.Chdir(filepath.Join("..", "example-go"))
								sess := git.StartPush(user, keyPath)
								os.Chdir(filepath.Join("..", "example-nodejs-express"))
								sess2 := git.StartPush(user, keyPath)
								Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
								Eventually(sess2, settings.MaxEventuallyTimeout).Should(Exit(0))
								git.Curl(app, "Powered by Deis")
								git.Curl(app2, "Powered by Deis")
							})

						})

					})

				})

			})

			Context("and who has a local git repo containing dockerfile source code", func() {

				BeforeEach(func() {
					output, err := cmd.Execute(`git clone https://github.com/deis/example-dockerfile-http.git`)
					Expect(err).NotTo(HaveOccurred(), output)
				})

				Context("and has run `deis apps:create` from within that repo", func() {

					var app model.App

					BeforeEach(func() {
						os.Chdir("example-dockerfile-http")
						app = apps.Create(user)
					})

					AfterEach(func() {
						apps.Destroy(user, app)
					})

					Specify("that user can deploy that app using a git push", func() {
						git.Push(user, keyPath, app, "Powered by Deis")
					})

					Specify("that user can deploy that app only once concurrently", func() {
						sess := git.StartPush(user, keyPath)
						// sleep for five seconds, then push the same app
						time.Sleep(5000 * time.Millisecond)
						sess2 := git.StartPush(user, keyPath)
						Eventually(sess2.Err).Should(Say("fatal: remote error: Another git push is ongoing"))
						Eventually(sess2).Should(Exit(128))
						git.PushUntilResult(user, keyPath,
							model.CmdResult{
								Out:      nil,
								Err:      []byte("Everything up-to-date"),
								ExitCode: 0,
							})
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
						git.Curl(app, "Powered by Deis")
					})

					Context("with a bad Dockerfile", func() {

						BeforeEach(func() {
							badCommit := `echo "BOGUS command" >> Dockerfile && EMAIL="ci@deis.com" git commit Dockerfile -m "Added a bogus command"`
							output, err := cmd.Execute(badCommit)
							Expect(err).NotTo(HaveOccurred(), output)
						})

						AfterEach(func() {
							undoCommit := `git reset --hard HEAD~`
							output, err := cmd.Execute(undoCommit)
							Expect(err).NotTo(HaveOccurred(), output)
						})

						Specify("that user can't deploy that app using a git push", func() {
							sess := git.StartPush(user, keyPath)
							Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(1))
							// TODO: this output doesn't show up 100% of the time! Needs a fix in dockerbuilder.
							// Eventually(sess.Err).Should(Say("Unknown instruction: BOGUS"))
							Eventually(sess.Err).Should(Say("error: failed to push some refs"))
						})

					})

					Context("and who has another local git repo containing dockerfile source code", func() {

						BeforeEach(func() {
							os.Chdir("..")
							output, err := cmd.Execute(`git clone https://github.com/deis/example-dockerfile-python.git`)
							Expect(err).NotTo(HaveOccurred(), output)
						})

						Context("and has run `deis apps:create` from within that repo", func() {

							var app2 model.App

							BeforeEach(func() {
								os.Chdir("example-dockerfile-python")
								app2 = apps.Create(user)
							})

							AfterEach(func() {
								apps.Destroy(user, app2)
							})

							Specify("that user can deploy both apps concurrently", func() {
								os.Chdir(filepath.Join("..", "example-dockerfile-http"))
								sess := git.StartPush(user, keyPath)
								os.Chdir(filepath.Join("..", "example-dockerfile-python"))
								sess2 := git.StartPush(user, keyPath)
								Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
								Eventually(sess2, settings.MaxEventuallyTimeout).Should(Exit(0))
								git.Curl(app, "Powered by Deis")
								git.Curl(app2, "Powered by Deis")
							})

						})

					})

				})

			})

		})

	})

})
