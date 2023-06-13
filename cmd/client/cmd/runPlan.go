package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vishu42/terrasome/pkg/targz"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func RunPlan(cmd *cobra.Command, args []string) {
	// create a temp tar file
	tempDir, err := os.MkdirTemp("", "terrasome")
	checkError(err)

	defer func() {
		// remove the temp directory
		cmd.Println("cleaning up...")
		cmd.Println("removing temp directory: " + tempDir)
		err := os.RemoveAll(tempDir)
		checkError(err)
	}()

	tempFile := tempDir + "/plan.tar"

	// log
	cmd.Println("created temp directory: " + tempDir)

	// create a .tar file
	file, err := os.Create(tempFile)
	checkError(err)
	defer file.Close()

	// get the current working directory
	cwd, err := os.Getwd()
	checkError(err)

	// tar the current working directory
	// log the current working dir
	cmd.Println("tarring current working directory: " + cwd)
	targz.Tardir(cwd, tempFile)
}
