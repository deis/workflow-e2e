package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	neturl "net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/onsi/ginkgo/reporters"
)

type Cmd struct {
	Env               []string
	CommandLineString string
}

type App struct {
	Name string
	URL  string
}

const (
	deisRouterServiceHost = "DEIS_ROUTER_SERVICE_HOST"
	deisRouterServicePort = "DEIS_ROUTER_SERVICE_PORT"
)

var (
	errMissingRouterHostEnvVar = fmt.Errorf("missing %s", deisRouterServiceHost)
	defaultMaxTimeout = 5 * time.Minute // gomega's default is 2 minutes
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandAppName() string {
	return fmt.Sprintf("test-%d", rand.Intn(999999999))
}

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)

	enableJunit := os.Getenv("JUNIT")
	if enableJunit == "true" {
		junitReporter := reporters.NewJUnitReporter("junit.xml")
		RunSpecsWithDefaultAndCustomReporters(t, "Deis Workflow", []Reporter{junitReporter})
	} else {
		RunSpecs(t, "Deis Workflow")
	}
}

var (
	randSuffix        = rand.Intn(1000)
	testUser          = fmt.Sprintf("test-%d", randSuffix)
	testPassword      = "asdf1234"
	testEmail         = fmt.Sprintf("test-%d@deis.io", randSuffix)
	testAdminUser     = "admin"
	testAdminPassword = "admin"
	testAdminEmail    = "admin@example.com"
	keyName           = fmt.Sprintf("deiskey-%v", randSuffix)
	url               = getController()
	debug             = os.Getenv("DEBUG") != ""
	homeHome          = os.Getenv("HOME")
)

var testRoot, testHome, keyPath, gitSSH string

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	// use the "deis" executable in the search $PATH
	output, err := exec.LookPath("deis")
	Expect(err).NotTo(HaveOccurred(), output)

	testHome, err = ioutil.TempDir("", "deis-workflow-home")
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("HOME", testHome)

	// register the test-admin user
	registerOrLogin(url, testAdminUser, testAdminPassword, testAdminEmail)

	// verify this user is an admin by running a privileged command
	sess, err := start("deis users:list")
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))

	sshDir := path.Join(testHome, ".ssh")

	// register the test user and add a key
	registerOrLogin(url, testUser, testPassword, testEmail)

	keyPath = createKey(keyName)

	// Write out a git+ssh wrapper file to avoid known_hosts warnings
	gitSSH = path.Join(sshDir, "git-ssh")
	sshFlags := "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
	if debug {
		sshFlags = sshFlags + " -v"
	}
	ioutil.WriteFile(gitSSH, []byte(fmt.Sprintf(
		"#!/bin/sh\nSSH_ORIGINAL_COMMAND=\"ssh $@\"\nexec /usr/bin/ssh %s -i %s \"$@\"\n",
		sshFlags, keyPath)), 0777)

	sess, err = start("deis keys:add %s.pub", keyPath)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Eventually(sess).Should(Say("Uploading %s.pub to deis... done", keyName))

	time.Sleep(5 * time.Second) // wait for ssh key to propagate
})

var _ = BeforeEach(func() {
	var err error
	var output string

	testRoot, err = ioutil.TempDir("", "deis-workflow-test")
	Expect(err).NotTo(HaveOccurred())

	os.Chdir(testRoot)
	output, err = execute(`git clone https://github.com/deis/example-go.git`)
	Expect(err).NotTo(HaveOccurred(), output)

	login(url, testUser, testPassword)
})

var _ = AfterEach(func() {
	err := os.RemoveAll(testRoot)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	os.Chdir(testHome)

	cancel(url, testUser, testPassword)
	cancel(url, testAdminUser, testAdminPassword)

	os.RemoveAll(fmt.Sprintf("~/.ssh/%s*", keyName))

	err := os.RemoveAll(testHome)
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("HOME", homeHome)
})

func register(url, username, password, email string) {
	sess, err := start("deis register %s --username=%s --password=%s --email=%s", url, username, password, email)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Say("Registered %s", username))
	Eventually(sess).Should(Say("Logged in as %s", username))
}

func registerOrLogin(url, username, password, email string) {
	sess, err := start("deis register %s --username=%s --password=%s --email=%s", url, username, password, email)

	Expect(err).To(BeNil())

	sess.Wait()

	if strings.Contains(string(sess.Err.Contents()), "must be unique") {
		// Already registered
		login(url, username, password)
	} else {
		Eventually(sess).Should(Exit(0))
		Eventually(sess).Should(SatisfyAll(
			Say("Registered %s", username),
			Say("Logged in as %s", username)))
	}
}

func cancel(url, username, password string) {
	// log in to the account
	login(url, username, password)

	// remove any existing test-* apps
	sess, err := start("deis apps")
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	re := regexp.MustCompile("test-.*")
	for _, app := range re.FindAll(sess.Out.Contents(), -1) {
		sess, err = start("deis destroy --app=%s --confirm=%s", app, app)
		Expect(err).To(BeNil())
		Eventually(sess).Should(Say("Destroying %s...", app))
		Eventually(sess).Should(Exit(0))
	}

	// cancel the account
	sess, err = start("deis auth:cancel --username=%s --password=%s --yes", username, password)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Eventually(sess).Should(Say("Account cancelled"))
}

func login(url, user, password string) {
	sess, err := start("deis login %s --username=%s --password=%s", url, user, password)
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Eventually(sess).Should(Say("Logged in as %s", user))
}

func logout() {
	sess, err := start("deis auth:logout")
	Expect(err).To(BeNil())
	Eventually(sess).Should(Exit(0))
	Eventually(sess).Should(Say("Logged out\n"))
}

// execute executes the command generated by fmt.Sprintf(cmdLine, args...) and returns its output as a cmdOut structure.
// this structure can then be matched upon using the SucceedWithOutput matcher below
func execute(cmdLine string, args ...interface{}) (string, error) {
	var cmd *exec.Cmd
	shCommand := fmt.Sprintf(cmdLine, args...)

	if debug {
		fmt.Println(shCommand)
	}

	cmd = exec.Command("/bin/sh", "-c", shCommand)
	outputBytes, err := cmd.CombinedOutput()

	output := string(outputBytes)

	if debug {
		fmt.Println(output)
	}

	return output, err
}

func start(cmdLine string, args ...interface{}) (*Session, error) {
	ourCommand := Cmd{Env: os.Environ(), CommandLineString: fmt.Sprintf(cmdLine, args...)}
	return startCmd(ourCommand)
}

func startCmd(command Cmd) (*Session, error) {
	cmd := exec.Command("/bin/sh", "-c", command.CommandLineString)
	cmd.Env = command.Env
	io.WriteString(GinkgoWriter, fmt.Sprintf("$ %s\n", command.CommandLineString))
	return Start(cmd, GinkgoWriter, GinkgoWriter)
}

func createKey(name string) string {
	keyPath := path.Join(testHome, ".ssh", name)
	os.MkdirAll(path.Join(testHome, ".ssh"), 0777)
	// create the key under ~/.ssh/<name> if it doesn't already exist
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		sess, err := start("ssh-keygen -q -t rsa -b 4096 -C %s -f %s -N ''", name, keyPath)
		Expect(err).To(BeNil())
		Eventually(sess).Should(Exit(0))
	}

	os.Chmod(keyPath, 0600)

	return keyPath
}

func getController() string {
	host := os.Getenv(deisRouterServiceHost)
	if host == "" {
		panicStr := fmt.Sprintf(`Set the router host and port for tests, such as:

$ %s=192.0.2.10 %s=31182 make test-integration`, deisRouterServiceHost, deisRouterServicePort)
		panic(panicStr)
	}
	// Make a xip.io URL if DEIS_ROUTER_SERVICE_HOST is an IP V4 address
	ipv4Regex := `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	matched, err := regexp.MatchString(ipv4Regex, host)
	if err != nil {
		panic(err)
	}
	if matched {
		host = fmt.Sprintf("deis.%s.xip.io", host)
	}
	port := os.Getenv(deisRouterServicePort)
	switch port {
	case "443":
		return "https://" + host
	case "80", "":
		return "http://" + host
	default:
		return fmt.Sprintf("http://%s:%s", host, port)
	}
}

// getRawRouter returns the URL to the deis router according to env vars.
//
// Returns an error if the minimal env vars are missing, or there was an error creating a URL from them.
func getRawRouter() (*neturl.URL, error) {
	host := os.Getenv(deisRouterServiceHost)
	if host == "" {
		return nil, errMissingRouterHostEnvVar
	}
	portStr := os.Getenv(deisRouterServicePort)
	switch portStr {
	case "443":
		return neturl.Parse(fmt.Sprintf("https://%s", host))
	case "80", "":
		return neturl.Parse(fmt.Sprintf("http://%s", host))
	default:
		return neturl.Parse(fmt.Sprintf("http://%s:%s", host, portStr))
	}
}

func createApp(name string, options ...string) *Session {
	cmd, err := start("deis apps:create %s %s", name, strings.Join(options, " "))
	Expect(err).NotTo(HaveOccurred())
	Eventually(cmd).Should(Say("created %s", name))
	Eventually(cmd).Should(Exit(0))

	return cmd
}

func destroyApp(app App) *Session {
	cmd, err := start("deis apps:destroy --app=%s --confirm=%s", app.Name, app.Name)
	Expect(err).NotTo(HaveOccurred())
	Eventually(cmd).Should(Exit(0))
	Eventually(cmd).Should(SatisfyAll(
		Say("Destroying %s...", app.Name),
		Say(`done in `)))

	return cmd
}

func deployApp(name string) App {
	os.Chdir(name)
	appName := getRandAppName()
	app := App{Name: appName, URL: strings.Replace(url, "deis", appName, 1)}

	cmd := createApp(app.Name)
	Eventually(cmd).Should(SatisfyAll(
		Say("Git remote deis added"),
		Say("remote available at ")))

	cmd, err := start("GIT_SSH=%s git push deis master", gitSSH)
	Expect(err).NotTo(HaveOccurred())
	Eventually(cmd.Err, "2m").Should(Say("Done, %s:v2 deployed to Deis", app.Name))
	Eventually(cmd).Should(Exit(0))

	return app
}

// cmdWithRetry runs the provided <cmd> repeatedly, once a second up to the
// supplied <timeout> until the <cmd> result contains the <expectedResult>
// An example use of this utility would be curl-ing a url and waiting
// until the response code matches the expected response
func cmdWithRetry(cmd Cmd, expectedResult string, timeout int) bool {
	var result string
	fmt.Printf("Waiting up to %d seconds for `%s` to return %s...\n", timeout, cmd, expectedResult)
	for i := 0; i < timeout; i++ {
		sess, err := startCmd(cmd)
		Expect(err).NotTo(HaveOccurred())
		result = string(sess.Wait().Out.Contents())
		if strings.Contains(result, expectedResult) {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("FAIL: '%s' does not match expected result of '%s'\n", result, expectedResult)
	return false
}

// gitInit simply invokes 'git init' and verifies the command is successful
func gitInit() {
	cmd, err := start("git init")
	Expect(err).NotTo(HaveOccurred())
	Eventually(cmd).Should(Say("Initialized empty Git repository"))
}

// gitClean destroys the .git directory and verifies the command is successful
func gitClean() {
	cmd, err := start("rm -rf .git")
	Expect(err).NotTo(HaveOccurred())
	Eventually(cmd).Should(Exit(0))
}
