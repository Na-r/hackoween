package main

import (
	"hack-o-ween-site/packages/auth"
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/countdown"
	"hack-o-ween-site/packages/debug"
	"hack-o-ween-site/packages/storage"
	"hack-o-ween-site/packages/settings"
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

	if file := filepath.Join(dir, path); fileExists(file) && filepath.Ext(file) != ".html" { // Serve a File
		//log.Println("Serving File")
		http.ServeFile(w, r, file)
	} else if file = filepath.Join(dir, path+".html"); fileExists(file) { // Serve templated HTML
		//log.Println("Serving HTML")
		key := cookie.GetCookie("session_key", r)
		m := make(map[string]interface{})
		session_key := ""
		if key != nil {
			session_key = key.(string)
			row := storage.GetAllFromTable_SessionKey(storage.AUTH_TABLE, session_key)
			if row != nil {
				var id int
				var auth_id, name, username, anon_name, pfp, session_key_filler, login_date string
				row.Scan(&id, &auth_id, &name, &username, &anon_name, &pfp, &session_key_filler, &login_date)

				m["Username"] = settings.GetUserName(session_key)
				m["PFP"] = pfp
				m["Login"] = auth.CheckExistingSession(r)
				m["Countdown"] = countdown.Get_duration()
			}
		}

		templates := getFilesInDir(templates_dir, ".html")
		templates = append(templates[:1], templates...)
		templates[0] = filepath.Join(dir, path+".html")

		tmpl := template.Must(template.ParseFiles(templates...))
		tmpl.Execute(w, m)

	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getFilesInDir(dir, ext string) (ret []string){
	files, _ := ioutil.ReadDir(templates_dir)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			ret = append(ret, filepath.Join(templates_dir, file.Name()))
		}
	}
	return
}

func main() {
	http.HandleFunc("/", HandleRequests)
	http.HandleFunc("/sign-out", auth.SignOutUser)
	http.HandleFunc("/oauth/github", auth.GithubAuthenticationRedirect)
	http.HandleFunc("/oauth/gitlab", auth.GitlabAuthenticationRedirect)
	http.HandleFunc("/oauth/google", auth.GoogleAuthenticationLogin)
	http.HandleFunc("/oauth/google/callback", auth.GoogleAuthenticationCallback)
	http.HandleFunc("/settings/save", settings.SaveSettings)

	http.HandleFunc("/debug", debug.DebugButton)
	http.ListenAndServe(":9956", nil)
}
