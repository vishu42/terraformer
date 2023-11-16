package pkg

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/vishu42/terraformer/pkg/github"
)

func CreateDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	// method must be POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the template id from the request body
	templateID := r.FormValue("templateID")

	// if template id is empty, return error
	if templateID == "" {
		http.Error(w, "template id cannot be empty", http.StatusBadRequest)
		return
	}

	// convert template id to int
	tidint, err := strconv.Atoi(templateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the template from the database
	template, err := GetTemplate(tidint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create a temp directory
	tempDir, err := CreateTempDir("terraformer")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// remove temp workdir after function execution is done
	defer RemoveTempDir(tempDir)

	// clone the template repository
	err = github.CloneRepo(template.GitRepository, tempDir, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// req envs
	// export ARM_CLIENT_ID
	// export ARM_CLIENT_SECRET
	// export ARM_TENANT_ID=
	// export ARM_SUBSCRIPTION_ID
	// execute terraform.sh
	fmt.Println("executing terraform.sh")
	cmd := exec.Command("terraform.sh", "plan", "plan.txt")
	cmd.Dir = tempDir
	armClientId := os.Getenv("ARM_CLIENT_ID")
	armClientSecret := os.Getenv("ARM_CLIENT_SECRET")
	armTenantId := os.Getenv("ARM_TENANT_ID")
	armSubscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")
	cmd.Env = []string{
		"ARM_CLIENT_ID=" + armClientId, "ARM_CLIENT_SECRET=" + armClientSecret,
		"ARM_TENANT_ID=" + armTenantId, "ARM_SUBSCRIPTION_ID=" + armSubscriptionId,
	}
	fw := FlushWriter{w: w}
	f, ok := w.(http.Flusher)
	if ok {
		fw.f = f
	}

	mw := io.MultiWriter(os.Stdout, &fw)
	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Start()
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
