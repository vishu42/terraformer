package impl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const (
	VersionEndpoint = "/version"
)

func RunVersion(cmd *cobra.Command, args []string) {
	// get the server address from the command line argument
	serverAddr, err := cmd.Flags().GetString("server-addr")
	if err != nil {
		log.Fatal(fmt.Errorf("error getting server address: %v", err))
	}

	// create a file in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("error getting home directory:", err)
	}

	// read ~/.terraformer file
	file, err := os.ReadFile(homeDir + "/.terraformer")
	if err != nil {
		log.Fatal("error reading ~/.terraformer file:", err)
	}

	// get the auth token from the file
	authToken := string(file)

	client := &http.Client{}

	// create a request
	req, err := http.NewRequest("GET", serverAddr+VersionEndpoint, nil)
	if err != nil {
		log.Fatal("error creating request:", err)
	}

	// add the auth token to the request header
	req.Header.Add("Authorization", "Bearer "+authToken)

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("error sending request:", err)
	}

	defer resp.Body.Close()

	// print the response body
	fmt.Println("response from server:")
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		log.Fatal("error printing response body:", err)
	}
}
