package impl

import (
	"log"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/vishu42/terraformer/pkg"
)

const (
	FileUploadEndpoint = "/plan"
)

type PlanOpts struct {
	// server address
	ServerAddr string
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

/*
ALGORITHM
- prepare a tar file of the current working directory
- send the tar file to the server
*/

func RunPlan(cmd *cobra.Command, args []string, o *PlanOpts) {
	// log svr
	cmd.Println("server address: " + o.ServerAddr)

	// quit
	log.Fatal("quitting")
	// create a temp tar file
	tempDir, err := os.MkdirTemp("", "terraformer")
	checkError(err)

	defer func() {
		// remove the temp directory
		cmd.Println("cleaning up...")
		cmd.Println("removing temp directory: " + tempDir)
		err := os.RemoveAll(tempDir)
		checkError(err)
	}()

	tarFile := tempDir + "/plan.tar.gz"

	// log
	cmd.Println("created temp directory: " + tempDir)

	// create a .tar.gz file
	file, err := os.Create(tarFile)
	checkError(err)
	defer file.Close()

	// get the current working directory
	cwd, err := os.Getwd()
	checkError(err)

	// log the current working dir
	cmd.Println("tarring current working directory: " + cwd)

	// tar the current working directory
	err = pkg.TarDir(cwd, tarFile)
	checkError(err)

	// get the server address from flags
	serverAddr, err := cmd.Flags().GetString("server-addr")
	if err != nil {
		checkError(err)
	}

	if serverAddr == "" {
		// get the server address from env
		log.Fatalf("server address not provided")
	}

	// log the server address
	cmd.Println("sending tar file to server: " + serverAddr)

	// send the tar file to the server
	uploadUrl, err := url.JoinPath(serverAddr, FileUploadEndpoint)
	checkError(err)
	err = pkg.UploadFile(tarFile, uploadUrl)
	checkError(err)
}
