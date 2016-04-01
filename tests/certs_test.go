package tests

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

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
func listCerts(profile string) *Session {
	sess, err := start("deis certs:list", profile)
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
	return sess
}

func removeCerts(profile string, certs []string) {
	for _, cert := range certs {
		sess, err := start("deis certs:remove %s", profile, cert)
		Eventually(sess).Should(Say("Removing %s...", cert))
		Eventually(sess).Should(Say("done"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Exit(0))
		Expect(err).NotTo(HaveOccurred())
	}

	Eventually(listCerts(profile)).Should(Say("No certs"))
}

func addCert(profile, certName, cert, key string) {
	sess, err := start("deis certs:add %s %s %s", profile, certName, cert, key)
	Eventually(sess).Should(Say("Adding SSL endpoint..."))
	Eventually(sess, defaultMaxTimeout).Should(Say("done"))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())

	Eventually(listCerts(profile).Wait().Out.Contents()).Should(ContainSubstring(certName))
}

func attachCert(profile, certName, domain string) {
	attachOrDetachCert(profile, certName, domain, "attach")
}

func detachCert(profile, certName, domain string) {
	attachOrDetachCert(profile, certName, domain, "detach")
}

func attachOrDetachCert(profile, certName, domain, attachOrDetach string) {
	// Explicitly build literal substring since 'domain'
	// may be a wildcard domain ('*.foo.com') and we don't want Gomega
	// interpreting this string as a regexp
	var substring string

	sess, err := start("deis certs:%s %s %s", profile, attachOrDetach, certName, domain)
	if attachOrDetach == "attach" {
		substring = fmt.Sprintf("Attaching certificate %s to domain %s...", certName, domain)
	} else {
		substring = fmt.Sprintf("Detaching certificate %s from domain %s...", certName, domain)
	}
	Eventually(sess.Wait().Out.Contents()).Should(ContainSubstring(substring))
	Eventually(sess, defaultMaxTimeout).Should(Say("done"))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

func certsInfo(profile, certName string) *Session {
	sess, err := start("deis certs:info %s", profile, certName)
	Eventually(sess).Should(Say("=== %s Certificate", certName))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())

	return sess
}

func verifySSLEndpoint(customSSLEndpoint, domain string, expectedStatusCode int) {
	cmdRetryTimeout := 60
	domain = strings.Replace(domain, "*", "blah", 1) // replace asterix if wildcard domain
	curlCmd := Cmd{CommandLineString: fmt.Sprintf(`curl -k -H "Host: %s" -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, customSSLEndpoint)}
	Eventually(cmdWithRetry(curlCmd, strconv.Itoa(expectedStatusCode), cmdRetryTimeout)).Should(BeTrue())
}

var _ = Describe("Certs", func() {
	var testApp App
	var domain string
	var certName string
	var certNames []string
	var customSSLEndpoint string
	var exampleRepo = "example-go"
	var testData TestData

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

	cleanUpCerts := func(profile string, certNames []string) {
		certsListing := string(listCerts(testData.Profile).Wait().Out.Contents()[:])
		if !strings.Contains(certsListing, "No certs") {
			removeCerts(profile, certNames)
		}
	}

	cleanUpDomains := func(profile string, domains []string) {
		for _, domain := range domains {
			removeDomain(profile, domain, testApp.Name)
		}
	}

	Context("with an app yet to be deployed", func() {
		BeforeEach(func() {
			testData = initTestData()
			gitInit()
			testApp = App{Name: getRandAppName()}
			createApp(testData.Profile, testApp.Name)
			domain = getRandDomain()
			certName = strings.Replace(domain, ".", "-", -1)
			certNames = []string{certName}
		})

		AfterEach(func() {
			cleanUpCerts(testData.Profile, certNames)
		})

		It("can add, attach, list, and remove certs", func() {
			addDomain(testData.Profile, domain, testApp.Name)
			addCert(testData.Profile, certName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)
			Eventually(certsInfo(testData.Profile, certName)).Should(Say("No connected domains"))
			attachCert(testData.Profile, certName, domain)
			Eventually(certsInfo(testData.Profile, certName)).Should(Say(domain))
		})
	})

	Context("with a deployed app", func() {

		BeforeEach(func() {
			testData = initTestData()
			os.Chdir(exampleRepo)
			appName := getRandAppName()
			createApp(testData.Profile, appName)
			testApp = deployApp(testData.Profile, appName)
			domain = getRandDomain()
			certName = strings.Replace(domain, ".", "-", -1)
			certNames = []string{certName}

			customSSLEndpoint = strings.Replace(testApp.URL, "http", "https", 1)
			portRegexp := regexp.MustCompile(`:\d+`)
			customSSLEndpoint = portRegexp.ReplaceAllString(customSSLEndpoint, "") // strip port
		})

		AfterEach(func() {
			defer os.Chdir("..")
			cleanUpCerts(testData.Profile, certNames)
		})

		It("can specify limit to number of certs returned by certs:list", func() {
			alternateCertName := strings.Replace(getRandDomain(), ".", "-", -1)
			certNames = append(certNames, alternateCertName)
			randDomainRegExp := `my-custom-[0-9]{0,9}-domain-com`

			addCert(testData.Profile, certName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)
			addCert(testData.Profile, alternateCertName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)

			sess, err := start("deis certs:list -l 0", testData.Profile)
			Eventually(sess).Should(Say("No certs"))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis certs:list --limit=1", testData.Profile)
			Eventually(sess).Should(Say(randDomainRegExp))
			Eventually(sess).Should(Not(Say(randDomainRegExp)))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())

			sess, err = start("deis certs:list", testData.Profile)
			Eventually(sess).Should(Say(randDomainRegExp))
			Eventually(sess).Should(Say(randDomainRegExp))
			Eventually(sess).Should(Exit(0))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can add, attach, list, and remove certs... improperly", func() {
			nonExistentCert := "non-existent.crt"
			nonExistentCertName := "non-existent-cert"

			addDomain(testData.Profile, domain, testApp.Name)

			// attempt to add cert with improper cert name (includes periods)
			sess, err := start("deis certs:add %s %s %s", testData.Profile, domain, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)
			Eventually(sess.Err).Should(Say("400 Bad Request"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to add cert with cert and key file swapped
			sess, err = start("deis certs:add %s %s %s", testData.Profile, certName, certs["wildcard"].KeyPath, certs["wildcard"].CertPath)
			Eventually(sess.Err).Should(Say("400 Bad Request"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to add cert with non-existent keys
			sess, err = start("deis certs:add %s %s %s", testData.Profile, certName, nonExistentCert, "non-existent.key")
			Eventually(sess.Err).Should(Say("open %s: no such file or directory", nonExistentCert))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to remove non-existent cert
			sess, err = start("deis certs:remove %s", testData.Profile, nonExistentCertName)
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to get info on non-existent cert
			sess, err = start("deis certs:info %s", testData.Profile, nonExistentCertName)
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to attach non-existent cert
			sess, err = start("deis certs:attach %s %s", testData.Profile, nonExistentCertName, domain)
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to detach non-existent cert
			sess, err = start("deis certs:detach %s %s", testData.Profile, nonExistentCertName, domain)
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			addCert(testData.Profile, certName, certs["wildcard"].CertPath, certs["wildcard"].KeyPath)

			// attempt to attach to non-existent domain
			sess, err = start("deis certs:attach %s %s", testData.Profile, certName, "non-existent-domain")
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to detach from non-existent domain
			sess, err = start("deis certs:detach %s %s", testData.Profile, certName, "non-existent-domain")
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			// attempt to remove non-existent cert
			sess, err = start("deis certs:remove %s", testData.Profile, nonExistentCertName)
			Eventually(sess.Err).Should(Say("404 Not Found"))
			Eventually(sess).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("multiple domains and certs", func() {
			domains := map[string]string{
				"wildcard": "*.foo.com",
				"foo":      "foo.com",
				"bar":      "bar.com",
			}
			domainNames := []string{domains["wildcard"], domains["foo"], domains["bar"]}

			AfterEach(func() {
				// need to cleanup domains as they are not named randomly as above
				cleanUpDomains(testData.Profile, domainNames)
			})

			It("can attach/detach 2 certs (1 wildcard) to/from 3 domains (1 wildcard)", func() {
				sharedCert := certs["wildcard"]
				certNames = []string{sharedCert.Name, certs["bar"].Name}

				// Add all 3 domains
				for _, domain := range domains {
					addDomain(testData.Profile, domain, testApp.Name)
				}

				// Add 2 certs
				addCert(testData.Profile, sharedCert.Name, sharedCert.CertPath, sharedCert.KeyPath)
				addCert(testData.Profile, certs["bar"].Name, certs["bar"].CertPath, certs["bar"].KeyPath)

				// Share wildcard cert betwtixt two domains, attach the other
				for _, domain := range []string{domains["wildcard"], domains["foo"]} {
					attachCert(testData.Profile, sharedCert.Name, domain)
				}
				attachCert(testData.Profile, certs["bar"].Name, domains["bar"])

				// With multiple strings to check, use substrings as ordering is non-deterministic
				// (Should(Say()) enforces strict ordering)
				bothDomains := fmt.Sprintf("%s,%s", domains["wildcard"], domains["foo"])
				Eventually(certsInfo(testData.Profile, sharedCert.Name).Wait().Out.Contents()).Should(ContainSubstring(bothDomains))
				Eventually(certsInfo(testData.Profile, certs["bar"].Name)).Should(Say(domains["bar"]))

				// All SSL endpoints should be good to go
				for _, domain := range domains {
					verifySSLEndpoint(customSSLEndpoint, domain, http.StatusOK)
				}

				// Detach shared cert from one domain and re-check endpoints
				detachCert(testData.Profile, sharedCert.Name, domains["wildcard"])
				Eventually(certsInfo(testData.Profile, sharedCert.Name)).Should(Say(domains["foo"]))
				verifySSLEndpoint(customSSLEndpoint, domains["wildcard"], http.StatusNotFound)
				verifySSLEndpoint(customSSLEndpoint, domains["foo"], http.StatusOK)

				detachCert(testData.Profile, certs["bar"].Name, domains["bar"])
				verifySSLEndpoint(customSSLEndpoint, domains["bar"], http.StatusNotFound)
			})

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
					domainNames = []string{domain}

					addDomain(testData.Profile, domain, testApp.Name)
					addCert(testData.Profile, certName, cert, key)
					attachCert(testData.Profile, certName, domain)
					Eventually(certsInfo(testData.Profile, certName).Wait().Out.Contents()).Should(ContainSubstring(domain))

					verifySSLEndpoint(customSSLEndpoint, domain, http.StatusOK)
					for _, otherDomain := range getOtherDomains(domain, domains) {
						verifySSLEndpoint(customSSLEndpoint, otherDomain, http.StatusNotFound)
					}

					detachCert(testData.Profile, certName, domain)

					Eventually(certsInfo(testData.Profile, certName)).Should(Say("No connected domains"))

					verifySSLEndpoint(customSSLEndpoint, domain, http.StatusNotFound)
				},

				Entry("a non-wildcard cert to a wildcard domain",
					domains["wildcard"], certs["www"].Name, certs["www"].CertPath, certs["www"].KeyPath),
				Entry("a non-wildcard cert to a non-wildcard domain",
					domains["foo"], certs["foo"].Name, certs["foo"].CertPath, certs["foo"].KeyPath),
				Entry("a non-wildcard cert to a non-wildcard domain",
					domains["bar"], certs["bar"].Name, certs["bar"].CertPath, certs["bar"].KeyPath),
			)
		})
	})
})
