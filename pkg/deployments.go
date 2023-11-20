package pkg

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/vishu42/terraformer/pkg/github"
)

/*

// deployment table schema
CREATE TABLE deployment (
    deployment_id INT AUTO_INCREMENT PRIMARY KEY,
    template_id INT,
    deployment_name VARCHAR(255),
    deployment_status VARCHAR(255),
    deployment_date DATETIME,
    deployment_type VARCHAR(255),
    waiting_for_approval BOOLEAN,
    plan_file VARCHAR(255),
    FOREIGN KEY (template_id) REFERENCES template(template_id)
);

// deployment_logs table schema
CREATE TABLE deployment_logs (
    deployment_id INT,
    deployment_log BLOB,
    FOREIGN KEY (deployment_id) REFERENCES deployment(deployment_id)
);

*/

const (
	StatusRunning         = "running"
	StatusStarted         = "started"
	StatusSuccess         = "success"
	StatusFailed          = "failed"
	TfActionPlan          = "plan"
	TfActionApply         = "apply"
	DeploymentTypePlanned = "planned"
	DeploymentTypeApplied = "applied"

	ErrAutoApproveMustBeTrue = "auto approve must be true"
)

func Plan(w http.ResponseWriter, r *http.Request, tid int64) (err error) {
	// get the template from the database
	template, err := GetTemplate(tid)
	if err != nil {
		http.Error(w, fmt.Errorf("error getting template - %v", err).Error(), http.StatusInternalServerError)
	}

	// create a temp directory
	tempDir, err := CreateTempDir("terraformer")
	if err != nil {
		return
	}

	// remove temp workdir after function execution is done
	defer RemoveTempDir(tempDir)

	// clone the template repository
	err = github.CloneRepo(template.GitRepository, tempDir, true)
	if err != nil {
		return
	}

	// generate a unique plan file name
	// get time now in unix nano
	// generate a random number
	randNo := rand.Intn(100)
	planFile := "plan" + strconv.FormatInt(tid, 10) + "." + strconv.FormatInt(int64(randNo), 10) + "." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".txt"

	// req envs
	// export ARM_CLIENT_ID
	// export ARM_CLIENT_SECRET
	// export ARM_TENANT_ID=
	// export ARM_SUBSCRIPTION_ID

	// prepare execution command
	cmd := exec.Command("terraform.sh", TfActionPlan, planFile)
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

	// add deployment to database with status as running
	db, err := sql.Open("mysql", SQL_DSN)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	defer db.Close()

	// Check the database connection
	err = db.Ping()
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	fmt.Println("checkpoint 1")
	deploymentName := template.TemplateName + strconv.FormatInt(tid, 10) + "." + strconv.FormatInt(int64(randNo), 10) + "." + strconv.FormatInt(time.Now().UnixNano(), 10)
	// create a deployment in the database
	res, err := db.Exec("INSERT INTO deployment (template_id, deployment_name, deployment_status, deployment_date, deployment_type, plan_file) VALUES (?, ?, ?, NOW(), ?, ?)", template.ID, deploymentName, StatusStarted, DeploymentTypePlanned, planFile)
	if err != nil {
		http.Error(w, fmt.Errorf("error inserting data into deployment - %v", err).Error(), http.StatusInternalServerError)
		return
	}

	// get the deployment id
	deploymentID, err := res.LastInsertId()
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	// get a writer to write to deployment_logs
	deploymentLogs := []byte{}
	deploymentLogsWriter := bytes.NewBuffer(deploymentLogs)

	// create a multiwriter to write to stdout and http response writer
	mw := io.MultiWriter(os.Stdout, &fw, deploymentLogsWriter)
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

	// get the cmd exit code
	exitCode := cmd.ProcessState.ExitCode()

	// update the deployment status in the database

	deploymentStatus := ""

	if exitCode == 0 {
		deploymentStatus = StatusSuccess
	} else {
		deploymentStatus = StatusFailed
	}

	waitingForApproval := false

	if deploymentStatus == StatusSuccess {
		waitingForApproval = true
	}
	_, err = db.Exec("UPDATE deployment SET deployment_status = ?, waiting_for_approval = ? WHERE deployment_id = ?", deploymentStatus, waitingForApproval, deploymentID)
	if err != nil {
		http.Error(w, fmt.Errorf("error updating deployment status - %v", err).Error(), http.StatusInternalServerError)
	}

	// update deployment_logs table
	_, err = db.Exec("INSERT INTO deployment_logs (deployment_id, deployment_log) VALUES (?, ?)", deploymentID, deploymentLogsWriter.String())
	if err != nil {
		http.Error(w, fmt.Errorf("error inserting data into deployment_logs - %v", err).Error(), http.StatusInternalServerError)
	}

	return
}

func AutoApproveApply(w http.ResponseWriter, r *http.Request, templateID int64) (err error) {
	// get the template from the database
	template, err := GetTemplate(templateID)
	if err != nil {
		http.Error(w, fmt.Errorf("error getting template - %v", err).Error(), http.StatusInternalServerError)
	}

	// create a temp directory
	tempDir, err := CreateTempDir("terraformer")
	if err != nil {
		return
	}

	// remove temp workdir after function execution is done
	defer RemoveTempDir(tempDir)

	// clone the template repository
	err = github.CloneRepo(template.GitRepository, tempDir, true)
	if err != nil {
		return
	}

	// generate a unique plan file name
	randNo := rand.Intn(100)
	planFile := "plan" + strconv.FormatInt(templateID, 10) + "." + strconv.FormatInt(int64(randNo), 10) + "." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".txt"

	// req envs
	// export ARM_CLIENT_ID
	// export ARM_CLIENT_SECRET
	// export ARM_TENANT_ID=
	// export ARM_SUBSCRIPTION_ID

	// prepare execution command
	cmd := exec.Command("terraform.sh", TfActionPlan, planFile)
	cmd.Dir = tempDir
	armClientId := os.Getenv("ARM_CLIENT_ID")
	armClientSecret := os.Getenv("ARM_CLIENT_SECRET")
	armTenantId := os.Getenv("ARM_TENANT_ID")
	armSubscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")
	cmd.Env = []string{
		"ARM_CLIENT_ID=" + armClientId, "ARM_CLIENT_SECRET=" + armClientSecret,
		"ARM_TENANT_ID=" + armTenantId, "ARM_SUBSCRIPTION_ID=" + armSubscriptionId,
		"AUTO_APPROVE=true",
	}
	fw := FlushWriter{w: w}
	f, ok := w.(http.Flusher)
	if ok {
		fw.f = f
	}

	// add deployment to database with status as running
	db, err := sql.Open("mysql", SQL_DSN)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	defer db.Close()

	// Check the database connection
	err = db.Ping()
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	res, err := db.Exec("INSERT INTO deployment (template_id, deployment_name, deployment_status, deployment_date, deployment_type, plan_file) VALUES (?, ?, ?, NOW(), ?, ?)", template.ID, template.TemplateName, StatusStarted, DeploymentTypeApplied, planFile)
	if err != nil {
		http.Error(w, fmt.Errorf("error inserting data into deployment - %v", err).Error(), http.StatusInternalServerError)
		return
	}

	// get the deployment id
	deploymentID, err := res.LastInsertId()

	// get a writer to write to deployment_logs
	deploymentLogs := []byte{}
	deploymentLogsWriter := bytes.NewBuffer(deploymentLogs)

	// create a multiwriter to write to stdout and http response writer
	mw := io.MultiWriter(os.Stdout, &fw, deploymentLogsWriter)
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

	// get the cmd exit code
	exitCode := cmd.ProcessState.ExitCode()

	// update the deployment status in the database

	deploymentStatus := ""

	if exitCode == 0 {
		deploymentStatus = StatusSuccess
	} else {
		deploymentStatus = StatusFailed
	}

	waitingForApproval := false

	_, err = db.Exec("UPDATE deployment SET deployment_status = ?, waiting_for_approval = ? WHERE deployment_id = ?", deploymentStatus, waitingForApproval, deploymentID)
	if err != nil {
		http.Error(w, fmt.Errorf("error updating deployment status - %v", err).Error(), http.StatusInternalServerError)
	}

	// update deployment_logs table
	_, err = db.Exec("INSERT INTO deployment_logs (deployment_id, deployment_log) VALUES (?, ?)", deploymentID, deploymentLogsWriter.String())
	if err != nil {
		http.Error(w, fmt.Errorf("error inserting data into deployment_logs - %v", err).Error(), http.StatusInternalServerError)
	}

	return
}

func Apply(w http.ResponseWriter, r *http.Request, deploymentId int64) (err error) {
	db, err := sql.Open("mysql", SQL_DSN)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	defer db.Close()

	// Check the database connection
	err = db.Ping()
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	// get the plan_file from database for the given deployment id from deployment table
	var planFile string

	err = db.QueryRow("SELECT plan_file FROM deployment WHERE deployment_id = ?", deploymentId).Scan(&planFile)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	// create a temp directory
	tempDir, err := CreateTempDir("terraformer")
	if err != nil {
		return
	}

	// remove temp workdir after function execution is done
	defer RemoveTempDir(tempDir)

	cmd := exec.Command("terraform.sh", TfActionPlan, planFile)
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
	// get a writer to write to deployment_logs
	deploymentLogs := []byte{}
	deploymentLogsWriter := bytes.NewBuffer(deploymentLogs)

	// create a multiwriter to write to stdout and http response writer
	mw := io.MultiWriter(os.Stdout, &fw, deploymentLogsWriter)
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

	// get the cmd exit code
	exitCode := cmd.ProcessState.ExitCode()

	// update the deployment status in the database

	deploymentStatus := ""

	if exitCode == 0 {
		deploymentStatus = StatusSuccess
	} else {
		deploymentStatus = StatusFailed
	}
	_, err = db.Exec("UPDATE deployment SET deployment_status = ?, deployment_logs = ?, deployment_type = ? WHERE deployment_id = ?", deploymentStatus, deploymentLogsWriter.String(), DeploymentTypeApplied, deploymentId)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	return
}

func ApplyDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	// method must be POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get deployment id from the request body
	deploymentID := r.FormValue("deploymentID")

	// convert deployment id to int
	didint, err := strconv.ParseInt(deploymentID, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("error parsing deploymentID - %v", err).Error(), http.StatusBadRequest)
		return
	}

	err = Apply(w, r, didint)
	if err != nil {
		http.Error(w, fmt.Errorf("error applying terraform - %v", err).Error(), http.StatusBadRequest)
		return
	}
}

func PlanDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	// method must be POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the template id from the request body
	templateID := r.FormValue("templateID")
	if templateID == "" {
		http.Error(w, "templateID cannot be empty", http.StatusBadRequest)
		return
	}

	// get the auto approve from the request body
	autoApprove := r.FormValue("autoApprove")
	if autoApprove == "" {
		autoApprove = "false"
	}

	// convert template id to int
	templateIDInt, err := strconv.ParseInt(templateID, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("error parsing templateID - %v", err).Error(), http.StatusBadRequest)
		return
	}

	if autoApprove == "true" {
		err := AutoApproveApply(w, r, templateIDInt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err := Plan(w, r, templateIDInt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
