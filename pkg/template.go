package pkg

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const (
	ErrGithubRepositoryEmpty = "github repository cannot be empty"
	SQL_DSN                  = "root:root@tcp(127.0.0.1:3306)/dev"
)

type Template struct {
	ID            int
	GitRepository string
}

func ListTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	// method must be GET
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templates, err := ListTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, t := range templates {
		fmt.Fprintf(w, "%d %s\n", t.ID, t.GitRepository)
	}
}

func CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// method must be POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get the github repository from the request body
	githubRepository := r.FormValue("githubRepository")

	// if github repository is empty, return error
	if githubRepository == "" {
		http.Error(w, "github repository cannot be empty", http.StatusBadRequest)
		return
	}

	err := CreateTemplate(githubRepository)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CreateTemplate(githubRepository string) error {
	// github repo can't be empty
	if githubRepository == "" {
		return fmt.Errorf(ErrGithubRepositoryEmpty)
	}

	// add the github repository to mysql database
	// Open a database connection
	db, err := sql.Open("mysql", SQL_DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check the database connection
	err = db.Ping()
	if err != nil {
		return err
	}

	// insert github repository to "templates" table
	_, err = db.Exec("INSERT INTO templates (git_repository) VALUES (?)", githubRepository)
	if err != nil {
		return err
	}

	return nil
}

func ListTemplates() (templates []Template, err error) {
	// Open a database connection
	db, err := sql.Open("mysql", SQL_DSN)
	if err != nil {
		return
	}
	defer db.Close()

	// Check the database connection
	err = db.Ping()
	if err != nil {
		return
	}

	// query the "templates" table
	rows, err := db.Query("SELECT * FROM templates")
	if err != nil {
		return
	}
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var t Template
		err = rows.Scan(&t.ID, &t.GitRepository)
		if err != nil {
			return
		}
		templates = append(templates, t)
	}

	return
}

func GetTemplate(templateId int) (template Template, err error) {
	// get template from the database
	// Open a database connection
	db, err := sql.Open("mysql", SQL_DSN)
	if err != nil {
		return
	}

	// Check the database connection
	err = db.Ping()
	if err != nil {
		return
	}

	// query the "templates" table
	err = db.QueryRow("SELECT * FROM templates WHERE template_id = ?", templateId).Scan(&template.ID, &template.GitRepository)
	if err != nil {
		return
	}

	return
}
