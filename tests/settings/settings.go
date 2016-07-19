package settings

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deis/workflow-e2e/tests/util"
)

const (
	DeisRootHostname = "k8s.local"
)

var (
	ActualHome               = os.Getenv("HOME")
	TestHome                 string
	TestRoot                 string
	DeisControllerURL        string
	DefaultEventuallyTimeout time.Duration
	MaxEventuallyTimeout     time.Duration
	GitSSH                   string
	Debug                    = os.Getenv("DEBUG") != ""
)

func init() {
	DeisControllerURL = getControllerURL()
	defaultEventuallyTimeoutStr := os.Getenv("DEFAULT_EVENTUALLY_TIMEOUT")
	if defaultEventuallyTimeoutStr == "" {
		DefaultEventuallyTimeout = 60 * time.Second
	} else {
		DefaultEventuallyTimeout, _ = time.ParseDuration(defaultEventuallyTimeoutStr)
	}

	maxEventuallyTimeoutStr := os.Getenv("MAX_EVENTUALLY_TIMEOUT")
	if maxEventuallyTimeoutStr == "" {
		MaxEventuallyTimeout = 600 * time.Second
	} else {
		MaxEventuallyTimeout, _ = time.ParseDuration(maxEventuallyTimeoutStr)
	}
}

func getControllerURL() string {
	// if DEIS_CONTROLLER_URL exists in the environment, use that
	controllerURL := os.Getenv("DEIS_CONTROLLER_URL")
	if controllerURL != "" {
		return controllerURL
	}

	// otherwise, rely on kubernetes and some DNS magic
	host := "deis." + DeisRootHostname
	if err := util.AddToEtcHosts(host); err != nil {
		log.Fatalf("Could not write %s to /etc/hosts (%s)", host, err)
	}
	// also gotta write a route for the builder
	builderHost := "deis-builder." + DeisRootHostname
	if err := util.AddToEtcHosts(builderHost); err != nil {
		log.Fatalf("Could not write %s to /etc/hosts (%s)", builderHost, err)
	}

	port := os.Getenv("DEIS_ROUTER_SERVICE_PORT")
	switch port {
	case "443":
		return "https://" + host
	case "80", "":
		return "http://" + host
	default:
		return fmt.Sprintf("http://%s:%s", host, port)
	}
}
