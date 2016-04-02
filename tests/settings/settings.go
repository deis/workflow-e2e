package settings

import (
	"os"
	"time"
)

var (
	ActualHome               = os.Getenv("HOME")
	TestHome                 string
	TestRoot                 string
	DeisControllerURL        = os.Getenv("DEIS_CONTROLLER_URL")
	DefaultEventuallyTimeout time.Duration
	MaxEventuallyTimeout     time.Duration
	GitSSH                   string
	Debug                    = os.Getenv("DEBUG") != ""
)

func init() {
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
