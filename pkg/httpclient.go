package pkg

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// type Client struct {
// 	// URL is the base URL for the API requests.
// 	URL *url.URL

// 	// HTTP client used to communicate with the API.
// 	client *http.Client

// 	// Logger is the logger used to log messages.
// 	Logger Logger

// 	// Debug specifies whether to log debug messages.
// 	Debug bool

// 	// ctx is the context used for requests.
// 	ctx context.Context
// }

// // NewClient returns a new API client.
// func NewClient(baseUrl string) (*Client, error) {
// 	baseURL, err := url.Parse(baseUrl)
// 	if err != nil {
// 		return nil, err
// 	}

// 	logger, err := New(false)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ctx := context.Background()

// 	// create new context with logger
// 	var loggerKey struct{}
// 	ctx = context.WithValue(ctx, loggerKey, logger)

// 	return &Client{
// 		URL:    baseURL,
// 		client: http.DefaultClient,
// 		Logger: logger,
// 		Debug:  false,
// 		ctx:    ctx,
// 	}, nil
// }

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
