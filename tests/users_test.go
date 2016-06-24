package tests

import (
	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis users", func() {

	Context("with an existing admin", func() {

		admin := model.Admin

		Specify("that admin can list all users", func() {
			sess, err := cmd.Start("deis users:list", &admin)
			Eventually(sess).Should(Say("=== Users"))
			output := string(sess.Out.Contents())
			Expect(output).To(ContainSubstring(admin.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
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

		Specify("that user cannot list all users", func() {
			sess, err := cmd.Start("deis users:list", &user)
			Eventually(sess.Err, settings.MaxEventuallyTimeout).Should(Say(util.PrependError(deis.ErrForbidden)))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})
	})

})
