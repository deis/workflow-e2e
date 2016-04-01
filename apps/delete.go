package apps

import (
	"io/ioutil"
	"os/exec"
)

// DeleteAll deletes all of the app names in appNames. It sends all successfully deleted apps on succCh and all deletion failures on errCh. After all deletion events have completed (with success or error), only doneCh will be closed. Note that succCh and errCh are never closed, to prevent accidental reception
func DeleteAll(appNames []Name, succCh chan<- Name, errCh chan<- error, doneCh chan<- struct{}) {
	for _, appName := range appNames {
		cmd := exec.Command("kubectl", "delete", "ns", appName.String())
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		if err := cmd.Run(); err != nil {
			errCh <- err
			continue
		}
		succCh <- appName
	}
	close(doneCh)
}
