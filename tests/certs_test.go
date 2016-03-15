package tests

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

type Cert struct {
	Name     string
	CertPath string
	KeyPath  string
}

// TODO: move to sister dir/package 'common'
//       for example, these could live in common/certs.go
// certs-specific common actions and expectations
func listCerts() *Session {
	sess, err := start("deis certs:list")
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))

	return sess
}

func removeCerts(certs []string) {
	for _, cert := range certs {
		sess, err := start("deis certs:remove %s", cert)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Say("Removing %s...", cert))
		Eventually(sess).Should(Say("done"))
		Eventually(sess).Should(Exit(0))
	}

	Eventually(listCerts()).Should(Say("No certs"))
}

func addCert(certName, cert, key string) {
	sess, err := start("deis certs:add %s %s %s", certName, cert, key)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("Adding SSL endpoint..."))
	Eventually(sess).Should(Say("done"))
	Eventually(sess).Should(Exit(0))

	Eventually(listCerts().Wait().Out.Contents()).Should(ContainSubstring(certName))
}

func attachCert(certName, domain string) {
	attachOrDetachCert(certName, domain, "attach")
}

func detachCert(certName, domain string) {
	attachOrDetachCert(certName, domain, "detach")
}

func attachOrDetachCert(certName, domain, attachOrDetach string) {
	// Explicitly build literal substring since 'domain'
	// may be a wildcard domain ('*.foo.com') and we don't want Gomega
	// interpreting this string as a regexp
	var substring string

	sess, err := start("deis certs:%s %s %s", attachOrDetach, certName, domain)
	Expect(err).NotTo(HaveOccurred())
	if attachOrDetach == "attach" {
		substring = fmt.Sprintf("Attaching certificate %s to domain %s...", certName, domain)
	} else {
		substring = fmt.Sprintf("Detaching certificate %s from domain %s...", certName, domain)
	}
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(substring))
	Eventually(sess).Should(Say("done"))
	Eventually(sess).Should(Exit(0))
}

func certsInfo(certName string) *Session {
	sess, err := start("deis certs:info %s", certName)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Say("=== %s Certificate", certName))
	Eventually(sess).Should(Exit(0))

	return sess
}

func verifySSLEndpoint(customSSLEndpoint, domain string, expectedStatusCode int) {
	maxRetryIterations := 15                         // ~1 iteration per second
	domain = strings.Replace(domain, "*", "blah", 1) // replace asterix if wildcard domain
	curlCmd := Cmd{CommandLineString: fmt.Sprintf(`curl -k -H "Host: %s" -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, customSSLEndpoint)}
	Eventually(cmdWithRetry(curlCmd, strconv.Itoa(expectedStatusCode), maxRetryIterations)).Should(BeTrue())
}

var _ = Describe("Certs", func() {
	var testApp App
	var domain string
	var certName string
	var certNames []string
	var customSSLEndpoint string
	var cleanup = true
	var exampleRepo = "example-go"

	certPath := path.Join(getDir(), "files/certs")
	certs := map[string]Cert{
		"www": Cert{
			Name:     "www-foo-com",
			CertPath: fmt.Sprintf("%s/www.foo.com.cert", certPath),
			KeyPath:  fmt.Sprintf("%s/www.foo.com.key", certPath)},
		"wildcard": Cert{
			Name:     "wildcard-foo-com",
			CertPath: fmt.Sprintf("%s/wildcard.foo.com.cert", certPath),
			KeyPath:  fmt.Sprintf("%s/wildcard.foo.com.key", certPath)},
		"foo": Cert{
			Name:     "foo-com",
			CertPath: fmt.Sprintf("%s/foo.com.cert", certPath),
			KeyPath:  fmt.Sprintf("%s/foo.com.key", certPath)},
		"bar": Cert{
			Name:     "bar-com",
			CertPath: fmt.Sprintf("%s/bar.com.cert", certPath),
			KeyPath:  fmt.Sprintf("%s/bar.com.key", certPath)},
	}

	cleanUpCerts := func(certNames []string) {
		certsListing := string(listCerts().Wait().Out.Contents()[:])
		if !strings.Contains(certsListing, "No certs") {
			removeCerts(certNames)
		}
	}

	Context("with an app yet to be deployed", func() {
		BeforeEach(func() {
			gitInit()
			testApp = App{Name: getRandAppName()}
			createApp(testApp.Name)
			domain = getRandDomain()
			certName = strings.Replace(domain, ".", "-", -1)
			certNames = []string{certName}
		})

		AfterEach(func() {
			if cleanup {
				cleanUpCerts(certNames)
			}
		})

		It("can add, attach, list, and remove certs", func() {
			addDomain(domain, testApp.Name)

			addCert(certName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)

			Eventually(certsInfo(certName)).Should(Say("No connected domains"))

			attachCert(certName, domain)

			Eventually(certsInfo(certName)).Should(Say(domain))
		})
	})

	Context("with a deployed app", func() {
		once := &sync.Once{}

		BeforeEach(func() {
			// Set up the test app only once and assume the suite will clean up.
			once.Do(func() {
				os.Chdir(exampleRepo)
				appName := getRandAppName()
				createApp(appName)
				testApp = deployApp(appName)
			})
			domain = getRandDomain()
			certName = strings.Replace(domain, ".", "-", -1)
			certNames = []string{certName}

			customSSLEndpoint = strings.Replace(testApp.URL, "http", "https", 1)
			portRegexp := regexp.MustCompile(`:\d+`)
			customSSLEndpoint = portRegexp.ReplaceAllString(customSSLEndpoint, "") // strip port
		})

		AfterEach(func() {
			defer os.Chdir("..")
			if cleanup {
				cleanUpCerts(certNames)
			}
		})

		It("can specify limit to number of certs returned by certs:list", func() {
			alternateCertName := strings.Replace(getRandDomain(), ".", "-", -1)
			certNames = append(certNames, alternateCertName)
			randDomainRegExp := `my-custom-[0-9]{0,9}-domain-com`

			addCert(certName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)
			addCert(alternateCertName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)

			sess, err := start("deis certs:list -l 0")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say("No certs"))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis certs:list --limit=1")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say(randDomainRegExp))
			Eventually(sess).Should(Not(Say(randDomainRegExp)))
			Eventually(sess).Should(Exit(0))

			sess, err = start("deis certs:list")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Say(randDomainRegExp))
			Eventually(sess).Should(Say(randDomainRegExp))
			Eventually(sess).Should(Exit(0))
		})

		It("can add, attach, list, and remove certs... improperly", func() {
			nonExistentCert := "non-existent.crt"
			nonExistentCertName := "non-existent-cert"

			addDomain(domain, testApp.Name)

			// attempt to add cert with improper cert name (includes periods)
			sess, err := start("deis certs:add %s %s %s", domain, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("400 Bad Request"))
			Eventually(sess).Should(Exit(1))

			// attempt to add cert with cert and key file swapped
			sess, err = start("deis certs:add %s %s %s", certName, certs["wildcard"].KeyPath, certs["wildcard"].CertPath)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("400 Bad Request"))
			Eventually(sess).Should(Exit(1))

			// attempt to add cert with non-existent keys
			sess, err = start("deis certs:add %s %s %s", certName, nonExistentCert, "non-existent.key")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("open %s: no such file or directory", nonExistentCert))
			Eventually(sess).Should(Exit(1))

			// attempt to remove non-existent cert
			sess, err = start("deis certs:remove %s", nonExistentCertName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			// attempt to get info on non-existent cert
			sess, err = start("deis certs:info %s", nonExistentCertName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			// attempt to attach non-existent cert
			sess, err = start("deis certs:attach %s %s", nonExistentCertName, domain)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			// attempt to detach non-existent cert
			sess, err = start("deis certs:detach %s %s", nonExistentCertName, domain)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			addCert(certName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)

			// attempt to attach to non-existent domain
			sess, err = start("deis certs:attach %s %s", certName, "non-existent-domain")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			// attempt to detach from non-existent domain
			sess, err = start("deis certs:detach %s %s", certName, "non-existent-domain")
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))

			// attempt to remove non-existent cert
			sess, err = start("deis certs:remove %s", nonExistentCertName)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
		})

		Context("multiple domains and certs", func() {

			domains := map[string]string{
				"wildcard": "*.foo.com",
				"foo":      "foo.com",
				"bar":      "bar.com",
			}

			getOtherDomains := func(myDomain string, domains map[string]string) []string {
				otherDomains := make([]string, 0, len(domains)-1)

				for _, domain := range domains {
					if domain != myDomain {
						otherDomains = append(otherDomains, domain)
					}
				}
				return otherDomains
			}

			DescribeTable("3 certs (no wildcards), 3 domains (1 wildcard)",

				func(domain, certName, cert, key string) {
					certNames = []string{certName}

					addDomain(domain, testApp.Name)

					addCert(certName, cert, key)

					attachCert(certName, domain)

					Eventually(certsInfo(certName).Wait().Out.Contents()).Should(ContainSubstring(domain))

					verifySSLEndpoint(customSSLEndpoint, domain, http.StatusOK)
					for _, otherDomain := range getOtherDomains(domain, domains) {
						verifySSLEndpoint(customSSLEndpoint, otherDomain, http.StatusNotFound)
					}

					detachCert(certName, domain)

					Eventually(certsInfo(certName)).Should(Say("No connected domains"))

					verifySSLEndpoint(customSSLEndpoint, domain, http.StatusNotFound)

					removeDomain(domain, testApp.Name)
				},

				Entry("a non-wildcard cert to a wildcard domain",
					domains["wildcard"], certs["www"].Name, certs["www"].CertPath, certs["www"].KeyPath),
				Entry("a non-wildcard cert to a non-wildcard domain",
					domains["foo"], certs["foo"].Name, certs["foo"].CertPath, certs["foo"].KeyPath),
				Entry("a non-wildcard cert to a non-wildcard domain",
					domains["bar"], certs["bar"].Name, certs["bar"].CertPath, certs["bar"].KeyPath),
			)

			DescribeTable("2 certs (1 wildcard), 3 domains (1 wildcard)",

				func(domain, certName, cert, key string) {
					// Explicitly build literal substrings since one of the domains
					// may be a wildcard domain ('*.foo.com') and we don't want Gomega
					// interpreting this string as a regexp
					var substring string
					var expectedStatus int
					cleanup = false

					addDomain(domain, testApp.Name)

					if domain != domains["foo"] { // use wildcard cert already added in prev run
						addCert(certName, cert, key)
					}

					attachCert(certName, domain)

					if domain == domains["foo"] { // verify wildcard domain also attached from prev run
						substring = fmt.Sprintf("%s,%s", domains["wildcard"], domains["foo"])
					} else {
						substring = domain
					}
					Eventually(certsInfo(certName).Wait().Out.Contents()).Should(ContainSubstring(substring))

					verifySSLEndpoint(customSSLEndpoint, domain, http.StatusOK)
					for _, otherDomain := range getOtherDomains(domain, domains) {
						// match wildcard cert still attached from prev run
						if domain == domains["foo"] && otherDomain == domains["wildcard"] {
							expectedStatus = http.StatusOK
						} else {
							expectedStatus = http.StatusNotFound
						}
						verifySSLEndpoint(customSSLEndpoint, otherDomain, expectedStatus)
					}

					if domain != domains["wildcard"] { // leave the cert attached for 'foo.com' domain
						detachCert(certName, domain)
						if domain == domains["foo"] { // need to also detach the wildcard cert since left attached above
							detachCert(certName, domains["wildcard"])
						}

						Eventually(certsInfo(certName)).Should(Say("No connected domains"))

						verifySSLEndpoint(customSSLEndpoint, domain, http.StatusNotFound)

						removeCerts([]string{certName})
					}
				},

				Entry("a wildcard cert to a wildcard domain",
					domains["wildcard"], certs["wildcard"].Name, certs["wildcard"].CertPath, certs["wildcard"].KeyPath),
				Entry("a wildcard cert to a non-wildcard domain",
					domains["foo"], certs["wildcard"].Name, certs["wildcard"].CertPath, certs["wildcard"].KeyPath),
				Entry("a non-wildcard cert to a non-wildcard domain",
					domains["bar"], certs["bar"].Name, certs["bar"].CertPath, certs["bar"].KeyPath),
			)
		})
	})
})
