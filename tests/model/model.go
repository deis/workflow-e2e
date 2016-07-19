package model

import (
	"bytes"
	"fmt"
	"math/rand"
	"path"
	"runtime"
	"strings"

	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"
)

var Admin = User{
	Username: "admin",
	Password: "admin",
	Email:    "admintest@deis.com",
}

type User struct {
	Username string
	Password string
	Email    string
}

func NewUser() User {
	randSuffix := rand.Intn(100000)
	return User{
		Username: fmt.Sprintf("test-%d", randSuffix),
		Password: "asdf1234",
		Email:    fmt.Sprintf("test-%d@deis.io", randSuffix),
	}
}

type App struct {
	Name string
	URL  string
}

func NewApp() App {
	name := fmt.Sprintf("test-%d", rand.Intn(999999999))
	app := App{
		Name: name,
		URL:  strings.Replace(settings.DeisControllerURL, "deis", name, 1),
	}
	// try adding the URL to /etc/hosts but don't cry if it's not in there because the user may
	// have other plans in store for DNS
	if err := util.AddToEtcHosts(fmt.Sprintf("%s.%s", name, settings.DeisRootHostname)); err != nil {
		fmt.Printf("WARNING: could not write %s to /etc/hosts (%s), continuing anyways\n",
			app.URL,
			err)
	}
	return app
}

type Cmd struct {
	Env               []string
	CommandLineString string
}

type Cert struct {
	Name     string
	CertPath string
	KeyPath  string
}

func NewCert() Cert {
	certPath := path.Join(getDir(), "..", "files", "certs")
	return Cert{
		Name:     getRandCertName(),
		CertPath: fmt.Sprintf("%s/www.foo.com.cert", certPath),
		KeyPath:  fmt.Sprintf("%s/www.foo.com.key", certPath),
	}
}

func getRandCertName() string {
	return fmt.Sprintf("%d-cert", rand.Intn(999999999))
}

func getDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

// CmdResult represents a generic command result, with expected Out, Err and
// ExitCode
type CmdResult struct {
	Out      []byte
	Err      []byte
	ExitCode int
}

// Satisfies returns whether or not the original CmdResult, ocd, meets all of
// the expectations contained in the expeced CmdResult, ecd
func (ocd CmdResult) Satisfies(ecd CmdResult) bool {
	if !bytes.Contains(ocd.Out, ecd.Out) {
		return false
	}
	if !bytes.Contains(ocd.Err, ecd.Err) {
		return false
	}
	if ocd.ExitCode != ecd.ExitCode {
		return false
	}
	return true
}

// String returns the CmdResult in printable form
func (ocd CmdResult) String() string {
	return fmt.Sprintf("[Out: '%s', Err: '%s', ExitCode: '%d']", ocd.Out, ocd.Err, ocd.ExitCode)
}
