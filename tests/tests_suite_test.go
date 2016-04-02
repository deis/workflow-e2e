package tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)

	enableJunit := os.Getenv("JUNIT")
	if enableJunit == "true" {
		junitReporter := reporters.NewJUnitReporter(filepath.Join(settings.ActualHome, fmt.Sprintf("junit-%d.xml", GinkgoConfig.ParallelNode)))
		RunSpecsWithDefaultAndCustomReporters(t, "Deis Workflow", []Reporter{junitReporter})
	} else {
		RunSpecs(t, "Deis Workflow")
	}
}

// SynchronizedBeforeSuite will run once and only once, even when tests are parallelized. It
// performs all the one-time setup required by the test suite.
var _ = SynchronizedBeforeSuite(func() []byte {
	// Verify the "deis" executable is on the $PATH
	output, err := exec.LookPath("deis")
	Expect(err).NotTo(HaveOccurred(), output)

	// Create temporary home directory for use by this test run.
	testHome, err := ioutil.TempDir("", "deis-workflow-home")
	Expect(err).NotTo(HaveOccurred())

	// When running parallel tests, Ginkgo seems to fork processes instead of using goroutines.
	// We set $HOME to our temporary home directory so it can be discovered and used by other
	// Ginkgo processes.
	os.Setenv("HOME", settings.TestHome)

	// Create and install a git wrapper script. This will allow us to always specify the private
	// key to use by means of an environment variable. This is a convenience that helps us run
	// tests in parallel, where each test user might have its own keys.
	sshHome := path.Join(testHome, ".ssh")
	os.MkdirAll(sshHome, 0777)
	settings.GitSSH = path.Join(sshHome, "git-ssh")
	sshFlags := "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
	if settings.Debug {
		sshFlags = sshFlags + " -v"
	}
	ioutil.WriteFile(settings.GitSSH, []byte(fmt.Sprintf("#!/bin/sh\nSSH_ORIGINAL_COMMAND=\"ssh $@\"\nexec /usr/bin/ssh %s -i \"$GIT_KEY\" \"$@\"\n", sshFlags)), 0777)

	// Set $HOME before we go any further. The user registration step below will need this to be
	// set correctly in order for the profile containing the user's auth token to be written to
	// the correct directory.
	os.Setenv("HOME", testHome)

	// Set the defaultEventuallyTimeout before we go any further. The next step carries out a user
	// registration and we'll want this timeout set before then.
	SetDefaultEventuallyTimeout(settings.DefaultEventuallyTimeout)

	// ATTEMPT to register the admin user. Since the FIRST user to regiser in a new cluster is
	// automatically the admin, it's vitally important that this happen now. If the admin user
	// already exists, this step will attempt to login as that user.
	auth.RegisterAdmin()

	// Return the value of testHome as bytes. Ginkgo will pass these to the function below, which
	// will be executed on every node (like BeforeSuite would if we were using it.)
	return []byte(testHome)
}, func(data []byte) {
	settings.TestHome = string(data)

	// Set $HOME for the benefit of all commands we will fork to execute.
	os.Setenv("HOME", settings.TestHome)

	// Derive settings.GitSSH from settings.TestHome.
	sshHome := path.Join(settings.TestHome, ".ssh")
	settings.GitSSH = path.Join(sshHome, "git-ssh")

	// Set the defaultEventuallyTimeout for ALL Ginko nodes.
	SetDefaultEventuallyTimeout(settings.DefaultEventuallyTimeout)
})

var _ = BeforeEach(func() {
	// Make a directory within the home directory for each test. This is to avoid collisions when
	// tests do things like clone git repos.
	var err error
	settings.TestRoot, err = ioutil.TempDir("", "deis-workflow-test")
	Expect(err).NotTo(HaveOccurred())
	// Everything we do, we do from within that directory...
	os.Chdir(settings.TestRoot)
	// But note that all test users and tests still share a common $HOME!
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	auth.CancelAdmin()
	os.RemoveAll(settings.TestHome)
})
