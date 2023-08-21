package pkg_test

import (
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/vishu42/terraformer/pkg"
)

func TestTarDir(t *testing.T) {
	dest := "../testfixture/testdir.tar.gz"
	src := "../testfixture/testdir"
	err := pkg.TarDir(src, dest)
	if err != nil {
		t.Errorf("TarDir() failed: %v", err)
	}

	cmd := exec.Command("tar", "-tvf", dest)
	output, err := cmd.Output()
	if err != nil {
		// if clone fails, the error will be of type *exec.ExitError
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Fatalf("%s", exitError.Stderr)
		}

		log.Fatalf("%s: %s", "error running tar command", err)
	}

	// assert if output includes dir1/dir2/file1.txt
	if !strings.Contains(string(output), "dir1/dir2/file1.txt") {
		t.Fatalf("want %s - got %s", "dir1/dir2/file1.txt", output)
	}
}

func TestUntarTar(t *testing.T) {
	err := pkg.UntarTar("../testfixture/testdirtargz", "../testfixture/testdir.tar.gz")
	if err != nil {
		t.Errorf("UntarTar() failed: %v", err)
	}

	cmd := exec.Command("ls", "-lart", "../testfixture/testdirtargz/dir1/dir2/file1.txt")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("%s: %s", "error running ls command", err)
	}
	// print output
	log.Println(string(output))
	// assert if output includes dir1/dir2/file1.txt
	if !strings.Contains(string(output), "dir1/dir2/file1.txt") {
		t.Fatalf("want %s - got %s", "dir1/dir2/file1.txt", output)
	}
}
