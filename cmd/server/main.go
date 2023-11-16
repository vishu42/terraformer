package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vishu42/terraformer/pkg"
)

func main() {
	config := pkg.LoadConfig()
	// t := pkg.Terraform{
	// 	Context: "",
	// 	Binary:  "terraform",
	// }
	// m.HandleFunc("/plan", t.Action)
	// m.HandleFunc("/apply", t.Action)
	// m.HandleFunc("/destroy", t.Action)

	m := mux.NewRouter()

	m.HandleFunc("/template", pkg.CreateTemplateHandler).Methods("POST")
	m.HandleFunc("/template", pkg.ListTemplatesHandler).Methods("GET")

	m.HandleFunc("/deployment", pkg.CreateDeploymentHandler).Methods("POST")

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
