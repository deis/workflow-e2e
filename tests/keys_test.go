package tests

import (
	"fmt"
	"math/rand"
	"time"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Keys", func() {
	It("can list and remove a key", func() {
		output, err := execute("deis keys:list")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("%s ssh-rsa", keyName))
	})

	It("can create and remove keys", func() {
		tempSSHKeyName := fmt.Sprintf("deiskey-%v", rand.Intn(1000))
		tempSSHKeyPath := createKey(tempSSHKeyName)

		sess, err := start("deis keys:add %s.pub", tempSSHKeyPath)
		Expect(err).To(BeNil())
		Eventually(sess).Should(Exit(0))
		Eventually(sess).Should(Say("Uploading %s.pub to deis... done", tempSSHKeyName))

		time.Sleep(5 * time.Second) // wait for ssh key to propagate

		output, err := execute("deis keys:remove %s", tempSSHKeyName)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("Removing %s SSH Key... done", tempSSHKeyName))

		os.RemoveAll(fmt.Sprintf("~/.ssh/%s*", tempSSHKeyName))
	})
})
