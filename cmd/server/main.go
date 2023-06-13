package main

import (
	"log"
	"net/http"
	"os/exec"
)

func main() {
	m := http.NewServeMux()

	m.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		// get binary version
		cmd := exec.Command("terraform", "version")
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		w.Write(stdoutStderr)
	})

	s := &http.Server{
		Addr:    ":80",
		Handler: m,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
