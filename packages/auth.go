package packages

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/random"
	_ "hack-o-ween-site/packages/storage"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"golang.org/x/oauth2/google"
)

const clientID_github = "e3608b8becc8be75c377"
const clientSecret_github = "edf9f8aea66771b6cd9077325c4c52799cb98710"

const clientID_gitlab = "09e978b08b2056b0e7d98fd8af90a55f14038189473c162d82fa97e4bfe1b608"
const clientSecret_gitlab = "c0022910a61977e536e6ad688723db34c538d2de549e99054db8d675c008b65f"

const clientID_google = "68153534942-ktejgnej2ki284h3c17ljm314998ah3r.apps.googleusercontent.com"
const clientSecret_google = "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"

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

type ErrorLevel uint8

const (
	Log ErrorLevel = iota
	Panic
	Fatal
)

func init() {

}

func ToSHA(str string) string {
	sha := sha256.Sum256([]byte(str))
	return string(sha[:])
}

func checkErr(err error, error_type ErrorLevel, desc string) {
	if err != nil {
		switch error_type {
		case Log:
			log.Printf("LOG | Error: %s\n%s", err.Error(), desc)

		case Panic:
			log.Panicf("PANIC | Error: %s\n%s", err.Error(), desc)

		case Fatal:
			log.Fatalf("FATAL | Error: %s\n%s", err.Error(), desc)

		}
	}
}

func GithubAuthenticationRedirect(w http.ResponseWriter, r *http.Request) {
	// First, we need to get the value of the `code` query param
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	code := r.FormValue("code")

	// Next, lets send the HTTP request to call the github oauth enpoint
	// to get our access token
	reqURL := fmt.Sprintf("%s?client_id=%s&client_secret=%s&code=%s",
		github.Endpoint.TokenURL, clientID_github, clientSecret_github, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	// Parse the request body into the `OAuthAccessResponse` struct
	var t OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	reqURL = "https://api.github.com/user"
	req, err = http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		fmt.Println("ERROR 1:", err)
		return
	}
	req.Header.Set("accept", "application/vnd.github.v3+json")
	req.Header.Set("authorization", "token "+t.AccessToken)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR 2:", err)
		return
	}

	var user_info map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&user_info)
	if err != nil {
		fmt.Println("ERROR 3:", err)
		return
	}

	if _, ok := user_info["id"]; !ok {
		log.Println("Error Retrieving GitHub User Info")
		return
	} else if _, ok := user_info["name"]; !ok {
		log.Println("Error Retrieving GitHub User Info")
		return
	} else if _, ok := user_info["login"]; !ok {
		log.Println("Error Retrieving GitHub User Info")
		return
	} else if _, ok := user_info["avatar_url"]; !ok {
		log.Println("Error Retrieving GitHub User Info")
		return
	}

	id := fmt.Sprintf("%.0f", user_info["id"].(float64)) + "GH"

	if CheckExistingUser(id) {
		LoginUser(id, w, r)
	} else {
		addNewUser(id, user_info["name"].(string),
			user_info["login"].(string), user_info["avatar_url"].(string), w, r)
	}
}

func GitlabAuthenticationRedirect(w http.ResponseWriter, r *http.Request) {
	fmt.Println("successful redirect to gitlab")
	// First, we need to get the value of the `code` query param
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	code := r.FormValue("code")

	// Next, lets send the HTTP request to call the github oauth enpoint
	// to get our access token
	reqURL := fmt.Sprintf("%s?client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=%s",
		gitlab.Endpoint.TokenURL, clientID_gitlab, clientSecret_gitlab, code, "http://localhost:9956/oauth/gitlab")
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	var access_info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&access_info); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	if _, ok := access_info["token_type"]; !ok {
		log.Println("Error Retrieving Gitlab Access Info")
		return
	} else if _, ok := access_info["access_token"]; !ok {
		log.Println("Error Retrieving Gitlab Access Info")
		return
	}

	reqURL = "https://gitlab.com/api/v4/user"
	req, err = http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		fmt.Println("ERROR 1:", err)
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", access_info["token_type"].(string)+" "+access_info["access_token"].(string))
	//req.Header.Set("access_token", access_info["access_token"].(string))
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ERROR 2:", err)
		return
	}

	var user_info map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&user_info)
	if err != nil {
		fmt.Println("ERROR 3:", err)
		return
	}

	if _, ok := user_info["id"]; !ok {
		log.Println("Error Retrieving GitLab User Info")
		return
	} else if _, ok := user_info["name"]; !ok {
		log.Println("Error Retrieving GitLab User Info")
		return
	} else if _, ok := user_info["username"]; !ok {
		log.Println("Error Retrieving GitLab User Info")
		return
	} else if _, ok := user_info["avatar_url"]; !ok {
		log.Println("Error Retrieving GitLab User Info")
		return
	}

	id := fmt.Sprintf("%.0f", user_info["id"].(float64)) + "GL"

	if CheckExistingUser(id) {
		LoginUser(id, w, r)
	} else {
		addNewUser(id, user_info["name"].(string),
			user_info["username"].(string), user_info["avatar_url"].(string), w, r)
	}
}

var google_oauthconf = &oauth2.Config{
	RedirectURL:  "http://localhost:9956/oauth/google/callback",
	ClientID:     clientID_google,
	ClientSecret: clientSecret_google,
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

func GoogleAuthenticationLogin(w http.ResponseWriter, r *http.Request) {

	state, err := random.GenerateRandomStringURLSafe(12)
	if err != nil {
		log.Println("Error in Google Authentication RNG")
		return
	}

	cookie.StoreCookie("google_oauthstate", state, w, r)

	url := google_oauthconf.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func GoogleAuthenticationCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState := cookie.GetCookie("google_oauthstate", r)

	if r.FormValue("state") != oauthState {
		log.Println("Invalid OAuth Google State")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	code := r.FormValue("code")

	token, err := google_oauthconf.Exchange(context.Background(), code)
	checkErr(err, Panic, "Failed Code Exchange")

	res, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	checkErr(err, Panic, "Failed GET on Google User Info")

	defer res.Body.Close()

	var user_info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&user_info); err != nil {
		checkErr(err, Panic, "Failed JSON Decode on Google User Info")
		return
	}

	if _, ok := user_info["id"]; !ok {
		log.Println("Error Retrieving Google User Info")
		return
	} else if _, ok := user_info["name"]; !ok {
		log.Println("Error Retrieving Google User Info")
		return
	} else if _, ok := user_info["picture"]; !ok {
		log.Println("Error Retrieving Google User Info")
		return
	}

	id := user_info["id"].(string) + "GG"

	if CheckExistingUser(id) {
		LoginUser(id, w, r)
	} else {
		addNewUser(id, user_info["name"].(string), "", user_info["picture"].(string), w, r)
	}
}

const AUTH_DATABASE = "./foo.db"

func addNewUser(id, name, username, pfp string, w http.ResponseWriter, r *http.Request) {
	session_key := generateSessionKey(id)
	login_date := strings.Split(time.Now().String(), " ")[0]

	log.Println("ID:", id)
	log.Println("Name:", name)
	log.Println("Username:", username)
	log.Println("Profile Pic:", pfp)
	log.Printf("Session Key: %x\n", session_key)
	log.Println("Login Date:", login_date)

	db, err := sql.Open("sqlite3", AUTH_DATABASE)
	checkErr(err, Fatal, "Failed Opening Auth Database")
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO foo (id, name, username, pfp, session_key, login_date) values(?, ?, ?, ?, ?, ?)")
	checkErr(err, Fatal, "Failed Transaction Preparation")
	defer stmt.Close()

	_, err = stmt.Exec(id, name, username, pfp, session_key, login_date)
	checkErr(err, Fatal, "Failed Transaction Execution")

	cookie.StoreCookie("session_key", session_key, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func generateSessionKey(id string) string {
	return ToSHA(id + strconv.FormatInt(time.Now().Unix(), 10))
}

func CheckExistingUser(id string) bool {
	db, err := sql.Open("sqlite3", AUTH_DATABASE)
	checkErr(err, Fatal, "Failed Opening Auth Database")
	defer db.Close()

	stmt, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM foo WHERE id = ?)")
	checkErr(err, Fatal, "Failed Transaction Preparation")
	defer stmt.Close()

	var exists int
	stmt.QueryRow(id).Scan(&exists)

	if exists == 1 {
		return true
	} else {
		return false
	}
}

func CheckExistingSession(r *http.Request) bool {
	session_key := cookie.GetCookie("session_key", r)
	if session_key == nil {
		return false
	}

	db, err := sql.Open("sqlite3", AUTH_DATABASE)
	checkErr(err, Fatal, "Failed Opening Auth Database")
	defer db.Close()

	stmt, err := db.Prepare("SELECT login_date FROM foo WHERE session_key = ?")
	checkErr(err, Fatal, "Failed Transaction Preparation")
	defer stmt.Close()

	var login_date_str string
	stmt.QueryRow(session_key).Scan(&login_date_str)

	temp := strings.Split(login_date_str, "-")
	temp_ints := []int{}
	for _, str := range temp {
		i, _ := strconv.Atoi(str)
		temp_ints = append(temp_ints, i)
	}

	if len(temp_ints) < 3 {
		return false
	}

	login_date := time.Date(temp_ints[0], time.Month(temp_ints[1]), temp_ints[2], 0, 0, 0, 0, time.Local)
	since := time.Since(login_date)

	if since.Hours() > time.Hour.Hours()*24*18 { // Session Expires in 18 Days
		return false
	} else {
		return true
	}
}

// "id" MUST be a valid user ID
func LoginUser(id string, w http.ResponseWriter, r *http.Request) {
	session_key := generateSessionKey(id)
	login_date := strings.Split(time.Now().String(), " ")[0]

	db, err := sql.Open("sqlite3", AUTH_DATABASE)
	checkErr(err, Fatal, "Failed Opening Auth Database")
	defer db.Close()

	stmt, err := db.Prepare("UPDATE foo SET session_key=?, login_date=? WHERE id=?")
	checkErr(err, Fatal, "Failed Transaction Preparation")
	defer stmt.Close()

	_, err = stmt.Exec(session_key, login_date, id)
	checkErr(err, Fatal, "Failed Transaction Execution")

	cookie.StoreCookie("session_key", session_key, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func SignOutUser(w http.ResponseWriter, r *http.Request) {
	cookie.StoreCookie("session_key", nil, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

/*

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(name)

	_, err = db.Exec("delete from foo")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		log.Fatal(err)
	}

	rows, err = db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return*/
