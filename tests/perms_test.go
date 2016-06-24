package tests

import (
	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/perms"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis perms", func() {

	Context("with an existing admin", func() {

		admin := model.Admin

		Specify("that admin can list admins", func() {
			sess, err := cmd.Start("deis perms:list --admin", &admin)
			Eventually(sess).Should(Say("=== Administrators"))
			Eventually(sess).Should(Say(admin.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Context("and another existing user", func() {

			var otherUser model.User

			BeforeEach(func() {
				otherUser = auth.Register()
			})

			AfterEach(func() {
				auth.Cancel(otherUser)
			})

			Specify("that admin can grant admin permissions to the other user", func() {
				sess, err := cmd.Start("deis perms:create %s --admin", &admin, otherUser.Username)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to system administrators... done\n", otherUser.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis perms:list --admin", &admin)
				Eventually(sess).Should(Say("=== Administrators"))
				Eventually(sess).Should(Say(otherUser.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("who owns an existing app", func() {

				var app model.App

				BeforeEach(func() {
					app = apps.Create(otherUser, "--no-remote")
				})

				AfterEach(func() {
					apps.Destroy(otherUser, app)
				})

				Specify("that admin can list permissions on the app owned by the second user", func() {
					sess, err := cmd.Start("deis perms:list --app=%s", &admin, app.Name)
					Eventually(sess).Should(Say("=== %s's Users", app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Context("and a third user also exists", func() {

					var thirdUser model.User

					BeforeEach(func() {
						thirdUser = auth.Register()
					})

					AfterEach(func() {
						auth.Cancel(thirdUser)
					})

					Specify("that admin can grant permissions on the app owned by the second user to the third user", func() {
						sess, err := cmd.Start("deis perms:create %s --app=%s", &admin, thirdUser.Username, app.Name)
						Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to %s collaborators... done\n", thirdUser.Username, app.Name))
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))

						sess, err = cmd.Start("deis perms:list --app=%s", &admin, app.Name)
						Eventually(sess).Should(Say("=== %s's Users", app.Name))
						Eventually(sess).Should(Say("%s", thirdUser.Username))
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))
					})

					Context("who has permissions on the second user's app", func() {

						BeforeEach(func() {
							sess, err := cmd.Start("deis perms:create %s --app=%s", &admin, thirdUser.Username, app.Name)
							Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to %s collaborators... done\n", thirdUser.Username, app.Name))
							Expect(err).NotTo(HaveOccurred())
							Eventually(sess).Should(Exit(0))
						})

						Specify("that admin can revoke the third user's permissions to an app owned by the second user", func() {
							sess, err := cmd.Start("deis perms:delete %s --app=%s", &admin, thirdUser.Username, app.Name)
							Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Removing %s from %s collaborators... done", thirdUser.Username, app.Name))
							Expect(err).NotTo(HaveOccurred())
							Eventually(sess).Should(Exit(0))

							sess, err = cmd.Start("deis perms:list --app=%s", &admin, app.Name)
							Eventually(sess).Should(Say("=== %s's Users", app.Name))
							Eventually(sess).ShouldNot(Say("%s", thirdUser.Username))
							Expect(err).NotTo(HaveOccurred())
							Eventually(sess).Should(Exit(0))
						})

					})

				})

			})

		})

		Context("and another existing admin", func() {

			var otherAdmin model.User

			BeforeEach(func() {
				otherAdmin = auth.Register()
				sess, err := cmd.Start("deis perms:create %s --admin", &admin, otherAdmin.Username)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to system administrators... done\n", otherAdmin.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			AfterEach(func() {
				auth.Cancel(otherAdmin)
			})

			Specify("the first admin can delete admin permissions from the second", func() {
				sess, err := cmd.Start("deis perms:delete %s --admin", &admin, otherAdmin.Username)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Removing %s from system administrators... done", otherAdmin.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis perms:list --admin", &admin)
				Eventually(sess).Should(Say("=== Administrators"))
				Expect(sess).ShouldNot(Say(otherAdmin.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

		})

	})

	Context("with an existing non-admin user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Specify("that user cannot list admin permissions", func() {
			sess, err := cmd.Start("deis perms:list --admin", &user)
			Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrForbidden)))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})

		Specify("that user cannot create admin permissions", func() {
			sess, err := cmd.Start("deis perms:create %s --admin", &user, user.Username)
			Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to system administrators...", user.Username))
			Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrForbidden)))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})

		Context("and an existing admin", func() {

			admin := model.Admin

			Specify("the non-admin user cannot delete the admin's admin permissions", func() {
				sess, err := cmd.Start("deis perms:delete %s --admin", &user, admin.Username)
				Eventually(sess.Err, settings.MaxEventuallyTimeout).Should(Say(util.PrependError(deis.ErrForbidden)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

		})

		Context("and an existing app belonging to that user", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list permissions for that app", func() {
				sess, err := cmd.Start("deis perms:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s's Users", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and another existing non-admin user also exists", func() {

				var otherUser model.User

				BeforeEach(func() {
					otherUser = auth.Register()
				})

				AfterEach(func() {
					auth.Cancel(otherUser)
				})

				Specify("that first user can grant permissions on that app to the second user", func() {
					perms.Create(user, app, otherUser)
					sess, err := cmd.Start("deis perms:list --app=%s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s's Users", app.Name))
					Eventually(sess).Should(Say("%s", otherUser.Username))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Context("who has already been granted permissions on that app", func() {

					BeforeEach(func() {
						perms.Create(user, app, otherUser)
					})

					Specify("that first user can list permissions for that app", func() {
						sess, err := cmd.Start("deis perms:list --app=%s", &user, app.Name)
						Eventually(sess).Should(Say("=== %s's Users", app.Name))
						Eventually(sess).Should(Say("%s", otherUser.Username))
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))
					})

					Specify("that first user can revoke permissions on that app", func() {
						perms.Delete(user, app, otherUser)
						sess, err := cmd.Start("deis perms:list --app=%s", &user, app.Name)
						Eventually(sess).Should(Say("=== %s's Users", app.Name))
						Eventually(sess).ShouldNot(Say("%s", otherUser.Username))
						Expect(err).NotTo(HaveOccurred())
						Eventually(sess).Should(Exit(0))
					})

				})

			})

		})

	})

})
