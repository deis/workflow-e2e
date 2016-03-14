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
func addDomain(domain, appName string) {
	addOrRemoveDomain(domain, appName, "add")
}

func removeDomain(domain, appName string) {
	addOrRemoveDomain(domain, appName, "remove")
}

func addOrRemoveDomain(domain, appName, addOrRemove string) {
	// Explicitly build literal substring since 'domain'
	// may be a wildcard domain ('*.foo.com') and we don't want Gomega
	// interpreting this string as a regexp
	var substring string

	sess, err := start("deis domains:%s %s --app=%s", addOrRemove, domain, appName)
	Expect(err).NotTo(HaveOccurred())
	if addOrRemove == "add" {
		substring = fmt.Sprintf("Adding %s to %s...", domain, appName)
	} else {
		substring = fmt.Sprintf("Removing %s from %s...", domain, appName)
	}
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(substring))
	Eventually(sess).Should(Say("done"))
	Eventually(sess).Should(Exit(0))
}

var _ = Describe("Domains", func() {
	var testApp App
	var domain string

	Context("with app yet to be deployed", func() {

		BeforeEach(func() {
			domain = getRandDomain()
			gitInit()

			testApp.Name = getRandAppName()
			createApp(testApp.Name)
		})

		AfterEach(func() {
			destroyApp(testApp)
			gitClean()
		})

		It("can list domains", func() {
			sess, err := start("deis domains:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", testApp.Name))
			Eventually(sess).Should(Exit(0))
		})

		It("can add and remove domains", func() {
			sess, err := start("deis domains:add %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Adding %s to %s...", domain, testApp.Name))
			Eventually(sess).Should(Say("done"))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis domains:remove %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Removing %s from %s...", domain, testApp.Name))
			Eventually(sess).Should(Say("done"))
		})
	})

	Context("with a deployed app", func() {
		var curlCmd Cmd
		var cmdRetryTimeout int

		BeforeEach(func() {
			cmdRetryTimeout = 10
			domain = getRandDomain()
			os.Chdir("example-go")
			appName := getRandAppName()
			createApp(appName)
			testApp = deployApp(appName)
		})

		AfterEach(func() {
			defer os.Chdir("..")
			destroyApp(testApp)
		})

		It("can add, list, and remove domains", func() {
			sess, err := start("deis domains:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", testApp.Name))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis domains:add %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Adding %s to %s...", domain, testApp.Name))
			Eventually(sess).Should(Say("done"))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis domains:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", domain))
			Eventually(sess).Should(Exit(0))

			// curl app at both root and custom domain, both should return http.StatusOK
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -H "Host: %s" -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())

			sess, err = start("deis domains:remove %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("Removing %s from %s...", domain, testApp.Name))
			Eventually(sess).Should(Say("done"))
			Eventually(sess).Should(Exit(0))

			// attempt to remove non-existent domain
			sess, err = start("deis domains:remove %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			sess, err = start("deis domains:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).Should(Say("%s", testApp.Name))
			Eventually(sess).Should(Not(Say("%s", domain)))
			Eventually(sess).Should(Exit(0))

			// curl app at both root and custom domain, custom should return http.StatusNotFound
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusOK), cmdRetryTimeout)).Should(BeTrue())
			curlCmd = Cmd{CommandLineString: fmt.Sprintf(`curl -sL -H "Host: %s" -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, testApp.URL)}
			Eventually(cmdWithRetry(curlCmd, strconv.Itoa(http.StatusNotFound), cmdRetryTimeout)).Should(BeTrue())
		})
	})
})
