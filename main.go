package main

import (
	"hack-o-ween-site/packages/auth"
	"hack-o-ween-site/packages/error_log"
	"hack-o-ween-site/packages/puzzle"
	"hack-o-ween-site/packages/settings"
	"hack-o-ween-site/packages/storage"
	"hack-o-ween-site/packages/utils"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func getTextInFile(fn string) string {
	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

var templates_dir = "templates/"

func HandleRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	dir := "public/"

	// Empty path defaults to index.html
	if path == "/" {
		path = "/index"
	}

	// log.Println("Request Handled:", path)

	if file := filepath.Join(dir, path); utils.FileExists(file) && filepath.Ext(file) != ".html" { // Serve a File
		//log.Println("Serving File")
		http.ServeFile(w, r, file)
	} else if file = filepath.Join(dir, path+".html"); utils.FileExists(file) { // Serve templated HTML
		//log.Println("Serving HTML")

		m := utils.GenUserTemplateData(r, path)

		templates := utils.GetFilesInDir(templates_dir, ".html")
		templates = append(templates[:1], templates...)
		templates[0] = filepath.Join(dir, path+".html")

		tmpl := template.Must(template.ParseFiles(templates...))
		tmpl.Execute(w, m)

	}

}

func main() {
	log_file, err := os.OpenFile("logs/" + utils.GetLogTimeString(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	error_log.CheckErr(err, error_log.Fatal, "Couldn't Open Log File")
	log.SetOutput(log_file)

	http.HandleFunc("/", HandleRequests)
	http.HandleFunc("/sign-out", auth.SignOutUser)
	http.HandleFunc("/oauth/github", auth.GithubAuthenticationRedirect)
	http.HandleFunc("/oauth/gitlab", auth.GitlabAuthenticationRedirect)
	http.HandleFunc("/oauth/google", auth.GoogleAuthenticationLogin)
	http.HandleFunc("/oauth/google/callback", auth.GoogleAuthenticationCallback)
	http.HandleFunc("/settings/save", settings.SaveSettings)

	puzzle.BindURLs("/alpha", storage.Alpha)

	// http.HandleFunc("/debug", debug.DebugButton)

	log.Println("Server Started!")
	http.ListenAndServe(":9956", nil)
}
