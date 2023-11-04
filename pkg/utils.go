package pkg

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-cmd/cmd"
)

func CreateTempDir(dirStub string) (tempDir string, err error) {
	// create a temp directory
	tempDir = "/tmp/" + dirStub + "-" + strconv.Itoa(rand.Intn(1000000))

	// create temp workdir
	err = MkDir(tempDir)
	if err != nil {
		return
	}

	return
}

// HandleStatus returns an error if the command failed to execute or there is a go error in status object
func HandleStatus(s cmd.Status) error {
	switch {
	case s.Error != nil:
		return s.Error
	case s.Exit != 0:
		err := fmt.Errorf("error while running %s\n%q", s.Cmd, s.Stderr)
		return err
	default:
		for _, line := range s.Stdout {
			fmt.Println(line)
		}
	}

	return nil
}

// remove temp workdir after function execution is done
func RemoveTempDir(tempDir string) {
	err := RmDir(tempDir)
	if err != nil {
		panic(err)
	}
}

// RmDir removes a directory
func RmDir(dir string) (err error) {
	rmdir := cmd.NewCmd("rm", "-rf", dir)
	status := <-rmdir.Start()
	err = HandleStatus(status)
	if err != nil {
		return err
	}
	return nil
}

// MkDir creates a directory
func MkDir(dir string) (err error) {
	mkdir := cmd.NewCmd("mkdir", dir)
	status := <-mkdir.Start()
	err = HandleStatus(status)
	if err != nil {
		return err
	}
	return nil
}
