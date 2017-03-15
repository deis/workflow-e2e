package tests

import (
	"fmt"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"strings"
)

var _ = Describe("deis limits", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.RegisterAndLogin()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app that has already been deployed", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
				builds.Create(user, app)
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can list that app's limits", func() {
				sess, err := cmd.Start("deis limits:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say(fmt.Sprintf("=== %s Limits", app.Name)))
				Eventually(sess).Should(Say("--- Memory\nUnlimited"))
				Eventually(sess).Should(Say("--- CPU\nUnlimited"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set a memory limit on that application", func() {
				sess, err := cmd.Start("deis limits:set cmd=64M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     64M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check that --memory also works
				// 128M
				sess, err = cmd.Start("deis limits:set --memory cmd=128M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     128M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check Kubernetes pods manifest
				sess, err = cmd.Start("HOME=%s kubectl get --all-namespaces pods -l app=%s --sort-by='.status.startTime' -o jsonpath={.items[*].spec.containers[0].resources}", nil, settings.ActualHome, app.Name)
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				resource := string(sess.Out.Contents())
				// try to get test latest pod, in case cmd still see terminated pod.
				// Also as per bug in https://github.com/kubernetes/kubernetes/issues/16707
				if strings.Contains(resource, "] map[") {
					resource = resource[strings.Index(resource, "] map[")+len("] "):]
				}
				Expect(resource).Should(SatisfyAny(
					Equal("map[requests:map[memory:128Mi] limits:map[memory:128Mi]]"),
					Equal("map[limits:map[memory:128Mi] requests:map[memory:128Mi]]")))

				// 0/100M
				sess, err = cmd.Start("deis limits:set cmd=0/100M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     0/100M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check Kubernetes pods manifest
				sess, err = cmd.Start("HOME=%s kubectl get --all-namespaces pods -l app=%s --sort-by='.status.startTime' -o jsonpath={.items[*].spec.containers[0].resources}", nil, settings.ActualHome, app.Name)
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				resource = string(sess.Out.Contents())
				// try to get test latest pod, in case cmd still see terminated pod.
				if strings.Contains(resource, "] map[") {
					resource = resource[strings.Index(resource, "] map[")+len("] "):]
				}
				Expect(resource).Should(SatisfyAny(
					Equal("map[requests:map[memory:0] limits:map[memory:100Mi]]"),
					Equal("map[limits:map[memory:100Mi] requests:map[memory:0]]")))

				// 50/100MB
				sess, err = cmd.Start("deis limits:set cmd=50M/100MB -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     50M/100M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check Kubernetes pods manifest
				sess, err = cmd.Start("HOME=%s kubectl get --all-namespaces pods -l app=%s --sort-by='.status.startTime' -o jsonpath={.items[*].spec.containers[0].resources}", nil, settings.ActualHome, app.Name)
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				resource = string(sess.Out.Contents())
				// try to get test latest pod, in case cmd still see terminated pod.
				if strings.Contains(resource, "] map[") {
					resource = resource[strings.Index(resource, "] map[")+len("] "):]
				}
				Expect(resource).Should(SatisfyAny(
					Equal("map[requests:map[memory:50Mi] limits:map[memory:100Mi]]"),
					Equal("map[limits:map[memory:100Mi] requests:map[memory:50Mi]]")))
			})

			Specify("that user can set a CPU limit on that application", func() {
				sess, err := cmd.Start("deis limits:set --cpu cmd=500m -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- CPU\ncmd     500m"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check Kubernetes pods manifest
				sess, err = cmd.Start("HOME=%s kubectl get --all-namespaces pods -l app=%s --sort-by='.status.startTime' -o jsonpath={.items[*].spec.containers[0].resources}", nil, settings.ActualHome, app.Name)
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				resource := string(sess.Out.Contents())
				// try to get test latest pod, in case cmd still see terminated pod.
				if strings.Contains(resource, "] map[") {
					resource = resource[strings.Index(resource, "] map[")+len("] "):]
				}
				Expect(resource).Should(SatisfyAny(
					Equal("map[requests:map[cpu:500m] limits:map[cpu:500m]]"),
					Equal("map[limits:map[cpu:500m] requests:map[cpu:500m]]")))
			})

			Specify("that user can unset a memory limit on that application", func() {
				// no memory has been set
				sess, err := cmd.Start("deis limits:unset cmd -a %s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))

				// Check that --memory also works
				sess, err = cmd.Start("deis limits:set --memory cmd=64M -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\ncmd     64M"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
				sess, err = cmd.Start("deis limits:unset --memory cmd -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("--- Memory\nUnlimited"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				// Check Kubernetes pods manifest
				sess, err = cmd.Start("HOME=%s kubectl get --all-namespaces pods -l app=%s -o jsonpath={.items[*].spec.containers[0].resources}", nil, settings.ActualHome, app.Name)
				Eventually(sess).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				// At least 1 pod have empty resources
				Expect(string(sess.Out.Contents())).Should(ContainSubstring("map[]"))
			})

			Specify("that user can unset a CPU limit on that application", func() {
				// no cpu has been set
				sess, err := cmd.Start("deis limits:unset --cpu cmd -a %s", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

		})

	})

})
