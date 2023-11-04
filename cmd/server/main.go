package main

import (
	"net/http"

	"github.com/vishu42/terraformer/pkg"
)

func main() {
	config := pkg.LoadConfig()
	t := pkg.Terraform{
		Context: "",
		Binary:  "terraform",
	}

	m := http.NewServeMux()

	m.HandleFunc("/version", t.Version)

	m.HandleFunc("/plan", t.Action)
	m.HandleFunc("/apply", t.Action)
	m.HandleFunc("/destroy", t.Action)

	m.HandleFunc("/new-template", pkg.CreateTemplateHandler)

	m.HandleFunc("/templates", pkg.ListTemplatesHandler)

	m.HandleFunc("/create-deployment", pkg.DeploymentPlanHandler)

	ea := pkg.NewEnsureAuth(&config, m)

	s := &http.Server{
		Addr:    ":80",
		Handler: ea,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
