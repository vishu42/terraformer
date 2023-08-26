package impl

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

const ClientId = "292def6c-3c4c-42ed-9ef2-8a197c7abc33"

func RunLogin(cmd *cobra.Command, args []string) {
	// Replace with your Azure App Registration configuration
	clientID := ClientId
	// clientSecret := "dlG8Q~n5mpzNhWQPGHMFFzTCtb-i6q0FNm066a50"
	redirectURL := "http://localhost:8000/callback"
	scopes := []string{"api://292def6c-3c4c-42ed-9ef2-8a197c7abc33/SCOPE1"} // Specify the required scopes

	conf := &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURL,
		Scopes:      scopes,
		Endpoint:    microsoft.AzureADEndpoint("4d16a70b-76a1-4ad7-944a-13513528982b"), // Replace with your Azure AD tenant ID
	}

	// Generate authorization URL
	authURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Printf("Open the following URL in your browser:\n\n%s\n\n", authURL)
	fmt.Println("After authentication, you will be redirected to the specified redirect URL.")

	m := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8000",
		Handler: m,
	}
	// Start a local HTTP server to listen for the callback
	m.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		// Exchange authorization code for access token
		token, err := conf.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal("Token exchange error:", err)
		}

		fmt.Println("Token:", token.AccessToken)

		// save the token in the config file

		// create a file in home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("error getting home directory:", err)
		}
		file, err := os.Create(homeDir + "/.terraformer")
		if err != nil {
			log.Fatal("error creating config file:", err)
		}
		defer file.Close()

		// write the token to the file
		_, err = file.WriteString(token.AccessToken)
		if err != nil {
			log.Fatal("error writing token to config file:", err)
		}

		// send the response
		fmt.Println("\nAuthentication successful! You can close the browser.")
		w.Write([]byte("Authentication successful! You can close the browser."))

		// Shutdown the server gracefully
		// TODO: figure out why browser doesn't show the response message?
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("HTTP server Shutdown error:", err)
		}
	})

	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			fmt.Println("HTTP server closed.")
		} else {
			log.Fatal("HTTP server ListenAndServe error:", err)
		}
	}
}
