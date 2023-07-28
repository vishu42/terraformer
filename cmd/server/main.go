package main

import (
	"net/http"

	"github.com/vishu42/terraformer/pkg/middleware"
	"github.com/vishu42/terraformer/pkg/terraform"
)

func main() {
	t := terraform.Terraform{
		Context: "",
		Binary:  "terraform",
	}

	m := http.NewServeMux()

	m.HandleFunc("/version", t.Version)

	m.HandleFunc("/plan", t.Action)
	m.HandleFunc("/apply", t.Action)
	m.HandleFunc("/destroy", t.Action)

	ea := middleware.NewEnsureAuth(m)

	s := &http.Server{
		Addr:    ":80",
		Handler: ea,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
