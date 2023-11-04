package github

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	GitBinary         = "git"
	ErrBinaryNotFound = "binary not found"
	ErrorRunningCmd   = "error running command"
)

// BinaryExists throws an error if the binary does not exist
func BinaryExists(binary string) (bool, error) {
	cmd := exec.Command("which", binary)
	_, err := cmd.Output()
	if err != nil {
		// if the binary does not exist, the error will be of type *exec.ExitError
		if exitError, ok := err.(*exec.ExitError); ok {
			return false, fmt.Errorf("%s", exitError.Stderr)
		}

		return false, fmt.Errorf("error running command: %s", err)
	}

	return true, nil
}

func ensureHTTPS(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return url
}

// CloneRepo clones a github repository
func CloneRepo(repo, workdir string, cloneInWorkDir bool) (err error) {
	// ensure HTTPS
	repo = ensureHTTPS(repo)

	ok, err := BinaryExists(GitBinary)
	if err != nil {
		return
	}
	if !ok {
		err = errors.New(ErrBinaryNotFound)
		return
	}
	opts := []string{"clone", repo}

	if cloneInWorkDir {
		opts = append(opts, ".")
	}
	// git clone
	cmd := exec.Command(GitBinary, opts...)
	cmd.Dir = workdir
	_, err = cmd.Output()
	if err != nil {
		// if clone fails, the error will be of type *exec.ExitError
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s", exitError.Stderr)
		}

		return fmt.Errorf("%s: %s", ErrorRunningCmd, err)
	}

	return
}
