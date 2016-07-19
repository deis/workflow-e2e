package util

import (
	"errors"
	"fmt"
	"os"
)

var errNoRouterHost = errors.New(`Set the router host and port for tests, such as:

$ DEIS_ROUTER_SERVICE_HOST=192.0.2.10 DEIS_ROUTER_SERVICE_PORT=31182 make test-integration`)

// PrependError adds 'Error: ' to an expected error, like the CLI does to error messages.
func PrependError(expected error) string {
	return "Error: " + expected.Error()
}

// AddToEtcHosts aliases the router IP address to the hostname via /etc/hosts
func AddToEtcHosts(hostname string) error {
	addr := os.Getenv("DEIS_ROUTER_SERVICE_HOST")
	if addr == "" {
		return errNoRouterHost
	}

	text := fmt.Sprintf("%s\t%s\n", addr, hostname)
	f, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(text)
	return err
}
