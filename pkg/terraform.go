package pkg

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/vishu42/terraformer/pkg/logger"
	"github.com/vishu42/terraformer/pkg/oauth"
)

type flushWriter struct {
	w io.Writer
	f http.Flusher
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}

	return
}

type Terraformer interface {
	Plan(w http.ResponseWriter, r *http.Request)
}

type Terraform struct {
	Context string
	Binary  string
}

func (t Terraform) Version(w http.ResponseWriter, r *http.Request) {
	oauthClaims, ok := oauth.FromContext(r.Context())
	if !ok {
		log.Fatal("error getting oauth claims from context")
	}

	// get logger from context
	logger, ok := logger.FromContext(r.Context())
	if !ok {
		log.Fatal("error getting logger from context")
	}

	auth := false
	for _, role := range oauthClaims.Roles {
		if role == "Terraformer.deployer" {
			auth = true
		}
	}

	if !auth {
		logger.Infof("unauthorized access to /version")

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cmd := exec.CommandContext(r.Context(), t.Binary, "version")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	w.Write(stdoutStderr)
}

func (t Terraform) TarUpload(w http.ResponseWriter, r *http.Request) (tempDir string, cleanTempDir func()) {
	// only accept POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// get the file
	formFile, _, err := r.FormFile("file")
	if err != nil {
		log.Fatal(err)
	}
	defer formFile.Close()

	// create temp directory to store the file
	tempDir, err = os.MkdirTemp("", "terraformer")
	if err != nil {
		log.Fatal(err)
	}

	tarFile, err := os.Create(tempDir + "/context.tar.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer tarFile.Close()

	// copy the file to the temp file
	_, err = io.Copy(tarFile, formFile)
	if err != nil {
		log.Fatal(err)
	}

	// untar
	err = UntarTar(tempDir, tempDir+"/context.tar.gz")
	if err != nil {
		log.Fatal(fmt.Errorf("error untarring tar file: %w", err))
	}

	// cleaup
	cleanTempDir = func() {
		// remove the temp directory
		log.Println("cleaning up...")
		log.Println("removing temp directory: " + tempDir)
		err := os.RemoveAll(tempDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (t Terraform) Action(w http.ResponseWriter, r *http.Request) {
	tempDir, cleanTempDir := t.TarUpload(w, r)
	t.Context = tempDir
	defer cleanTempDir()

	fw := flushWriter{w: w}
	f, ok := w.(http.Flusher)
	if ok {
		fw.f = f
	}

	arg := ""

	switch {
	case r.URL.Path == "/plan":
		arg = "plan"
	case r.URL.Path == "/apply":
		arg = "apply"
	case r.URL.Path == "/destroy":
		arg = "destroy"
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("BAD REQUEST"))
		return
	}

	cmd := exec.Command(t.Binary, arg)

	mw := io.MultiWriter(os.Stdout, &fw)
	cmd.Stdout = mw
	cmd.Dir = t.Context

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
}
