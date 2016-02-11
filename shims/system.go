package shims

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Shim struct {
	OutFile  *os.File
	ShimFile *os.File
}

func CreateSystemShim(toShim string) (Shim, error) {
	tempDir := strings.TrimSuffix(os.TempDir(), "/")
	// create out file for shim to write to
	outFile, err := ioutil.TempFile(tempDir, fmt.Sprintf("%s.out", toShim))
	if err != nil {
		return Shim{}, err
	}

	// create shim file
	shimLogic := []byte(fmt.Sprintf("#!/bin/sh\necho $@ > %s", outFile.Name()))
	shimFileName := fmt.Sprintf("%s/%s", tempDir, toShim)
	shimFile, err := os.Create(shimFileName)
	if err != nil {
		return Shim{}, err
	}
	if _, err = shimFile.Write(shimLogic); err != nil {
		return Shim{}, err
	}
	shimFile.Chmod(0777)

	return Shim{OutFile: outFile, ShimFile: shimFile}, nil
}

func RemoveShim(shim Shim) {
	os.Remove(shim.OutFile.Name())
	os.Remove(shim.ShimFile.Name())
}

func SubstituteEnvVar(env []string, envKey string, envValue string) []string {
	// create clone of env provided
	newEnv := make([]string, len(env))
	copy(newEnv, env)

	// find and delete key/value in question
	for i, e := range newEnv {
		pair := strings.Split(e, "=")
		if pair[0] == envKey {
			newEnv = append(newEnv[:i], newEnv[i+1:]...)
		}
	}
	// substitute new key/value
	newEnv = append(newEnv, fmt.Sprintf("%s=%s", envKey, envValue))

	return newEnv
}

func PrependPath(env []string, prefix string) []string {
	path := os.Getenv("PATH")
	return SubstituteEnvVar(env, "PATH", fmt.Sprintf("%s:%s", prefix, path))
}
