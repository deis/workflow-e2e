package model

import (
	"fmt"
	"math/rand"
	"path"
	"runtime"
	"strings"

	"github.com/deis/workflow-e2e/tests/settings"
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
	return App{
		Name: name,
		URL:  strings.Replace(settings.DeisControllerURL, "deis", name, 1),
	}
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
