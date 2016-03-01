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

var _ = Describe("Domains", func() {
	var testApp App
	var domain string

	Context("with app yet to be deployed", func() {

		BeforeEach(func() {
			domain = getRandDomain()
			gitInit()

			testApp.Name = getRandAppName()
			cmd := createApp(testApp.Name)
			Eventually(cmd).Should(SatisfyAll(
				Say("Git remote deis added"),
				Say("remote available at ")))
		})

		AfterEach(func() {
			destroyApp(testApp)
			gitClean()
		})

		It("can list domains", func() {
			sess, err := start("deis domains:list --app=%s", testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("=== %s Domains", testApp.Name))
			Eventually(sess).ShouldNot(Say("%s", testApp.Name))
			Eventually(sess).Should(Exit(0))
		})

		It("cannot add or remove domains", func() {
			sess, err := start("deis domains:add %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("500 Internal Server Error")) // better error/explanation needed from cli
			Eventually(sess).Should(Exit(1))

			sess, err = start("deis domains:remove %s --app=%s", domain, testApp.Name)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("500 Internal Server Error")) // better error/explanation needed from cli
			Eventually(sess).Should(Exit(1))
		})
	})

	Context("with a deployed app", func() {
		var curlCmd Cmd
		var cmdRetryTimeout int

		BeforeEach(func() {
			cmdRetryTimeout = 10
			domain = getRandDomain()
			testApp = deployApp("example-go")
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
			Eventually(sess.Err).Should(Say("500 Internal Server Error")) // better error/explanation needed from cli
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
