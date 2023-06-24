package main

import (
	"net/http"

	"github.com/vishu42/terrasome/pkg/terraform"
)

func main() {
	t := terraform.Terraform{
		Context: "",
		Binary:  "terraform",
	}

	m := http.NewServeMux()

	m.HandleFunc("/version", t.Version)

	m.HandleFunc("/plan", t.Action)

	s := &http.Server{
		Addr:    ":80",
		Handler: m,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
