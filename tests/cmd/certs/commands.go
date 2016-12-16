package certs

import (
	"errors"
	"fmt"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var ErrNoCertMatch = errors.New("\"No Certificate matches the given query.\"")

// The functions in this file implement SUCCESS CASES for commonly used `deis certs` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// List executes `deis certs:list` as the specified user.
func List(user model.User) *Session {
	sess, err := cmd.Start("deis certs:list", &user)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	return sess
}

// Add executes `deis certs:add` as the specified user to add the specified cert.
func Add(user model.User, cert model.Cert) {
	sess, err := cmd.Start("deis certs:add %s %s %s", &user, cert.Name, cert.CertPath, cert.KeyPath)
	Eventually(sess).Should(Say("Adding SSL endpoint..."))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	Eventually(List(user).Wait().Out.Contents()).Should(ContainSubstring(cert.Name))
}

// Remove executes `deis certs:remove` as the specified user to remove the specified cert.
func Remove(user model.User, cert model.Cert) {
	sess, err := cmd.Start("deis certs:remove %s", &user, cert.Name)
	Eventually(sess).Should(Say("Removing %s...", cert.Name))
	Eventually(sess).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	Eventually(List(user).Wait().Out.Contents()).ShouldNot(ContainSubstring(cert.Name))
}

// Attach executes `deis certs:attach` as the specified user to attach the specified cert to the
// specified domain.
func Attach(user model.User, cert model.Cert, domain string) {
	sess, err := cmd.Start("deis certs:attach %s %s", &user, cert.Name, domain)
	// Explicitly build literal substring since 'domain' may be a wildcard domain ('*.foo.com') and
	// we don't want Gomega interpreting this string as a regexp
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(fmt.Sprintf("Attaching certificate %s to domain %s...", cert.Name, domain)))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
}

// Detatch executes `deis certs:detach` as the specified user to detach the specified cert from
// the specified domain.
func Detach(user model.User, cert model.Cert, domain string) {
	sess, err := cmd.Start("deis certs:detach %s %s", &user, cert.Name, domain)
	// Explicitly build literal substring since 'domain' may be a wildcard domain ('*.foo.com') and
	// we don't want Gomega interpreting this string as a regexp
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(fmt.Sprintf("Detaching certificate %s from domain %s...", cert.Name, domain)))
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
}

// Info executes `deis certs:info` as the specified user to retrieve information about the
// specified cert.
func Info(user model.User, cert model.Cert) *Session {
	sess, err := cmd.Start("deis certs:info %s", &user, cert.Name)
	Eventually(sess).Should(Say("=== %s Certificate", cert.Name))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	return sess
}
