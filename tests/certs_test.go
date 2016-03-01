package tests

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
	// . "github.com/onsi/gomega/gbytes"
	// . "github.com/onsi/gomega/gexec"
)

var _ = Describe("Certs", func() {

	// Basic "Smoke" test
	XIt("can add, list, and remove certs", func() {
		// "deis domains:add %s --app=%s", domain, app
		// "deis certs:list", app
		// "deis certs:add %s %s", certPath, keyPath
		// wait for 60 seconds until cert generation is done?
		// curl the custom SSL endpoint
		// "deis certs:remove %s", domain
		// "deis certs:list", app
		// curl the custom SSL endpoint, should fail
		// curl app at both root and custom domain, custom should fail
		// "deis domains:remove %s --app=%s", domain, app
	})

	// See examples in https://github.com/deis/workflow/tree/master/rootfs/api/tests
	// for other use cases perhaps needing to be tested
})
