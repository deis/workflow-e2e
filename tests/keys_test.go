package tests

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Keys", func() {
	var testData TestData

	BeforeEach(func() {
		testData = initTestData()
	})

	It("can list and remove a key", func() {
		sess, err := start("deis keys:list", testData.Profile)
		Eventually(sess, defaultMaxTimeout).Should(Say(fmt.Sprintf("%s ssh-rsa", testData.KeyName)))
		Eventually(sess).Should(Exit(0))
		Expect(err).NotTo(HaveOccurred())
	})

	It("can create and remove keys", func() {
		tempSSHKeyName := fmt.Sprintf("deiskey-%v", rand.Intn(1000))
		tempSSHKeyPath := createKey(testData.Username, tempSSHKeyName)

		sess, err := start("deis keys:add %s.pub", testData.Profile, tempSSHKeyPath)
		Eventually(sess, defaultMaxTimeout).Should(Say("Uploading %s.pub to deis... done", tempSSHKeyName))
		Eventually(sess).Should(Exit(0))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(5 * time.Second) // wait for ssh key to propagate

		sess, err = start("deis keys:remove %s", testData.Profile, tempSSHKeyName)
		Eventually(sess, defaultMaxTimeout).Should(Say("Removing %s SSH Key... done", tempSSHKeyName))
		Eventually(sess).Should(Exit(0))
		Expect(err).NotTo(HaveOccurred())

		os.RemoveAll(fmt.Sprintf("~/.ssh/%s*", tempSSHKeyName))
	})
})
