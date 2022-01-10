package main

import (
	"hack-o-ween-site/packages"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type HTMLData struct {
	Login      bool
	Username   string
	Countdown  string
	HTMLheader template.HTML
}

func HandleRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	dir := "public/"

	//log.Println("PATH:", path)

	header := template.HTML(
		`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Hack-o-Ween</title>
		<link rel="stylesheet" href="style.css">
	</head>`)

	// Empty path defaults to index.html
	if path == "/" {
		path = "/index"
	}

	if file := filepath.Join(dir, path); fileExists(file) && filepath.Ext(file) != ".html" { // Serve a File
		//log.Println("Serving File")
		http.ServeFile(w, r, file)
	} else if file = filepath.Join(dir, path+".html"); fileExists(file) { // Serve templated HTML
		//log.Println("Serving HTML")

		data := HTMLData{
			Login:      packages.CheckExistingSession(r),
			Username:   "testificate",
			Countdown:  packages.Get_duration(),
			HTMLheader: header,
		}

		tmpl := template.Must(template.ParseFiles(filepath.Join(dir, path+".html")))
		tmpl.Execute(w, data)

	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {
	http.HandleFunc("/", HandleRequests)
	http.HandleFunc("/sign-out", packages.SignOutUser)
	http.HandleFunc("/oauth/github", packages.GithubAuthenticationRedirect)
	http.HandleFunc("/oauth/gitlab", packages.GitlabAuthenticationRedirect)
	http.HandleFunc("/oauth/google", packages.GoogleAuthenticationLogin)
	http.HandleFunc("/oauth/google/callback", packages.GoogleAuthenticationCallback)
	//ackages.GoogleAuthenticationRedirect()
	http.ListenAndServe(":9956", nil)
}
