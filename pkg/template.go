package pkg

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

/*
template table schema
# template table
CREATE TABLE template (
	template_id INT AUTO_INCREMENT PRIMARY KEY,
	git_repository VARCHAR(255)
	template_name VARCHAR(255)
);
*/

const (
	ErrGithubRepositoryEmpty = "github repository cannot be empty"
)

type Template struct {
	ID            int
	GitRepository string
	TemplateName  string
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
		fmt.Fprintf(w, "%d %s %s\n", t.ID, t.GitRepository, t.TemplateName)
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
		http.Error(w, "githubRepository cannot be empty", http.StatusBadRequest)
		return
	}

	// get the templateName from the request body
	templateName := r.FormValue("templateName")

	// if templateName is empty, return error
	if templateName == "" {
		http.Error(w, "templateName cannot be empty", http.StatusBadRequest)
		return
	}

	err := CreateTemplate(githubRepository, templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CreateTemplate(githubRepository, templateName string) error {
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

	// insert github repository to "template" table
	_, err = db.Exec("INSERT INTO template (git_repository, template_name) VALUES (?, ?)", githubRepository, templateName)
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
	rows, err := db.Query("SELECT * FROM template")
	if err != nil {
		return
	}
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var t Template
		err = rows.Scan(&t.ID, &t.GitRepository, &t.TemplateName)
		if err != nil {
			return
		}
		templates = append(templates, t)
	}

	return
}

func GetTemplate(templateId int64) (template Template, err error) {
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
	err = db.QueryRow("SELECT * FROM template WHERE template_id = ?", templateId).Scan(&template.ID, &template.GitRepository, &template.TemplateName)
	if err != nil {
		return
	}

	return
}
