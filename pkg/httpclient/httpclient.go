package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func UploadFile(filePath string, url string) error {
	// log the url
	fmt.Println("runPlan.go: uploading file to: " + url)

	// Open the tar file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a buffer to store the file contents
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	// Create a form field for the file
	fileField, err := mw.CreateFormFile("file", file.Name())
	if err != nil {
		return err
	}

	// Copy the file contents to the form field
	_, err = io.Copy(fileField, file)
	if err != nil {
		return err
	}

	// Close the writer to finalize the multipart form
	mw.Close()

	// Create a POST request with the file
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	// Set the Content-Type header to the multipart form data
	req.Header.Set("Content-Type", mw.FormDataContentType())

	// Send the request
	client := &http.Client{
		Timeout: time.Minute * 5,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return err
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("file upload failed with status code: %d", resp.StatusCode)
	}

	return nil
}
