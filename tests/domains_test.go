package tests

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func getRandDomain() string {
	return fmt.Sprintf("my-custom-%d.domain.com", rand.Intn(999999999))
}

// TODO: move to sister dir/package 'common'
//       for example, these could live in common/domains.go
// demains-specific common actions and expectations
func addDomain(profile string, domain, appName string) {
	addOrRemoveDomain(profile, domain, appName, "add")
}

func removeDomain(profile string, domain, appName string) {
	addOrRemoveDomain(profile, domain, appName, "remove")
}

func addOrRemoveDomain(profile string, domain, appName, addOrRemove string) {
	// Explicitly build literal substring since 'domain'
	// may be a wildcard domain ('*.foo.com') and we don't want Gomega
	// interpreting this string as a regexp
	var substring string

	sess, err := start("deis domains:%s %s --app=%s", profile, addOrRemove, domain, appName)
	if addOrRemove == "add" {
		substring = fmt.Sprintf("Adding %s to %s...", domain, appName)
	} else {
		substring = fmt.Sprintf("Removing %s from %s...", domain, appName)
	}
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(substring))
	Eventually(sess, defaultMaxTimeout).Should(Say("done"))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("Domains", func() {
	var testApp App
	var domain string
	var testData TestData

	Context("with app yet to be deployed", func() {

		BeforeEach(func() {
			testData = initTestData()
			domain = getRandDomain()
			gitInit()

			testApp.Name = getRandAppName()
			createApp(testData.Profile, testApp.Name)
		})

		It("can list domains", func() {
			sess, err := start("deis domains:list --app=%s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", testApp.Name))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can add and remove domains", func() {
			sess, err := start("deis domains:add %s --app=%s", testData.Profile, domain, testApp.Name)
			Eventually(sess).Should(Say("Adding %s to %s...", domain, testApp.Name))
			Eventually(sess, defaultMaxTimeout).Should(Say("done"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis domains:remove %s --app=%s", testData.Profile, domain, testApp.Name)
			Eventually(sess).Should(Say("Removing %s from %s...", domain, testApp.Name))
			Eventually(sess, defaultMaxTimeout).Should(Say("done"))
		})
	})

	Context("with a deployed app", func() {
		var curlCmd Cmd
		var cmdRetryTimeout int
		var testData TestData

		BeforeEach(func() {
			testData = initTestData()
			cmdRetryTimeout = 60
			domain = getRandDomain()
			os.Chdir("example-go")
			appName := getRandAppName()
			createApp(testData.Profile, appName)
			testApp = deployApp(testData.Profile, appName)
		})

		AfterEach(func() {
			defer os.Chdir("..")
		})

		It("can add, list, and remove domains", func() {
			sess, err := start("deis domains:list --app=%s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", testApp.Name))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis domains:add %s --app=%s", testData.Profile, domain, testApp.Name)
			Eventually(sess).Should(Say("Adding %s to %s...", domain, testApp.Name))
			Eventually(sess, defaultMaxTimeout).Should(Say("done"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis domains:list --app=%s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", domain))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			// curl app at both root and custom domain, both should return http.StatusOK
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -H "Host: %s" -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

			sess, err = start("deis domains:remove %s --app=%s", testData.Profile, domain, testApp.Name)
			Eventually(sess).Should(Say("Removing %s from %s...", domain, testApp.Name))
			Eventually(sess, defaultMaxTimeout).Should(Say("done"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			// attempt to remove non-existent domain
			sess, err = start("deis domains:remove %s --app=%s", testData.Profile, domain, testApp.Name)
			Eventually(sess.Err, defaultMaxTimeout).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			sess, err = start("deis domains:list --app=%s", testData.Profile, testApp.Name)
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", testApp.Name))
			Eventually(sess).Should(Not(Say("%s", domain)))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			// curl app at both root and custom domain, custom should return http.StatusNotFound
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -H "Host: %s" -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusNotFound), cmdRetryTimeout)).Should(BeTrue())
		})
	})
})
