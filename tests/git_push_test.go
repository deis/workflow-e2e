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
			user = auth.RegisterAndLogin()
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

					Context("with a bad buildpack", func() {

						BeforeEach(func() {
							badBuildpackURL := "https://github.com/deis/heroku-buildpack-epic-fail.git"
							sess, err := cmd.Start("deis config:set BUILDPACK_URL=%s", &user, badBuildpackURL)
							Expect(err).NotTo(HaveOccurred())
							Eventually(sess).Should(Say("BUILDPACK_URL"))
							Eventually(sess).Should(Exit(0))
						})

						AfterEach(func() {
							sess, err := cmd.Start("deis config:unset BUILDPACK_URL", &user)
							Expect(err).NotTo(HaveOccurred())
							Eventually(sess).ShouldNot(Say("BUILDPACK_URL"))
							Eventually(sess).Should(Exit(0))
						})

						Specify("that user can't deploy the app", func() {
							sess := git.StartPush(user, keyPath)
							Eventually(sess.Err, settings.MaxEventuallyTimeout).Should(Say("-----> Fetching custom buildpack"))
							Eventually(sess.Err).Should(Say("exited with code 1, stopping build"))
							Eventually(sess).Should(Exit(1))
						})

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

					Specify("and can execute deis run successfully", func() {
						git.Push(user, keyPath, app, "Powered by Deis")
						sess, err := cmd.Start("deis run env -a %s", &user, app.Name)
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
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

					Specify("that user can deploy that app using a git push after setting config values", func() {
						sess, err := cmd.Start("deis config:set -a %s PORT=80 POWERED_BY=midi-chlorians", &user, app.Name)
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Say("Creating config"))
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
						output := string(sess.Out.Contents())
						Expect(output).To(MatchRegexp(`PORT\s+80`))
						Expect(output).To(MatchRegexp(`POWERED_BY\s+midi-chlorians`))
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))

						git.Push(user, keyPath, app, "Powered by midi-chlorians")
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
							Eventually(sess.Err).Should(Say("Unknown instruction: BOGUS"))
							Eventually(sess.Err).Should(Say("error: failed to push some refs"))
						})

					})

					Context("and who has another local git repo containing dockerfile source code", func() {

						BeforeEach(func() {
							os.Chdir("..")
							output, err := cmd.Execute(`git clone https://github.com/deis/example-dockerfile-procfile-http.git`)
							Expect(err).NotTo(HaveOccurred(), output)
						})

						Context("and has run `deis apps:create` from within that repo", func() {

							var app2 model.App

							BeforeEach(func() {
								os.Chdir("example-dockerfile-procfile-http")
								app2 = apps.Create(user)
							})

							AfterEach(func() {
								apps.Destroy(user, app2)
							})

							Specify("that user can deploy both apps concurrently", func() {
								os.Chdir(filepath.Join("..", "example-dockerfile-http"))
								sess := git.StartPush(user, keyPath)
								os.Chdir(filepath.Join("..", "example-dockerfile-procfile-http"))
								sess2 := git.StartPush(user, keyPath)
								Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
								Eventually(sess2, settings.MaxEventuallyTimeout).Should(Exit(0))
								git.Curl(app, "Powered by Deis")
								git.Curl(app2, "Powered by Deis")
								_ = listProcs(user, app2, "web")
							})

						})

					})

					Specify("and can execute deis run successfully", func() {
						git.Push(user, keyPath, app, "Powered by Deis")
						sess, err := cmd.Start("deis run env -a %s", &user, app.Name)
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
					})

				})

			})

		})

	})

})
