package main

import (
	"net/http"

	"github.com/vishu42/terraformer/pkg"
)

func main() {
	t := pkg.Terraform{
		Context: "",
		Binary:  "terraform",
	}

	m := http.NewServeMux()

	m.HandleFunc("/version", t.Version)

	m.HandleFunc("/plan", t.Action)
	m.HandleFunc("/apply", t.Action)
	m.HandleFunc("/destroy", t.Action)

	ea := pkg.NewEnsureAuth(m)

	s := &http.Server{
		Addr:    ":80",
		Handler: ea,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
