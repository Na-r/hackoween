package main

import (
	"net/http"

	_ "golang.org/x/oauth2/github"
	_ "golang.org/x/oauth2/gitlab"
	_ "golang.org/x/oauth2/google"
)

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.ListenAndServe(":9956", nil)
}
