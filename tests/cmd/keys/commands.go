package keys

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// The functions in this file implement SUCCESS CASES for commonly used `deis keys` subcommands.
// This allows each of these to be re-used easily in multiple contexts.

// Add executes `deis keys:add` as the specified user to add a new key to that user's account.
func Add(user model.User) (string, string) {
	keyName, keyPath := createKey()
	sess, err := cmd.Start("deis keys:add %s.pub", &user, keyPath)
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Uploading %s.pub to deis... done", keyName))
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	time.Sleep(5 * time.Second) // Wait for the key to propagate before continuing
	return keyName, keyPath
}

// Remove executes `deis keys:remove` as the specified user to remove the specified key from that
// user's account.
func Remove(user model.User, keyName string) {
	sess, err := cmd.Start("deis keys:remove %s", &user, keyName)
	Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Removing %s SSH Key... done", keyName))
	Eventually(sess).Should(Exit(0))
	Expect(err).NotTo(HaveOccurred())
}

func createKey() (string, string) {
	keyName := fmt.Sprintf("deiskey-%v", rand.Intn(1000))
	sshHome := path.Join(settings.TestHome, ".ssh")
	os.MkdirAll(sshHome, 0777)
	keyPath := path.Join(sshHome, keyName)
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		_, err := cmd.Execute("ssh-keygen -q -t rsa -b 4096 -C %s -f %s -N ''", keyName, keyPath)
		Expect(err).NotTo(HaveOccurred())
	}
	os.Chmod(keyPath, 0600)
	return keyName, keyPath
}
