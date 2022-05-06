package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/random"
	"hack-o-ween-site/packages/storage"
	"hack-o-ween-site/packages/error_log"
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

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

func ToSHA(str string) string {
	sha := sha256.Sum256([]byte(str))
	return string(sha[:])
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
		github.Endpoint.TokenURL, storage.PRIVATE_DATA.GH.ID, storage.PRIVATE_DATA.GH.SECRET, code)
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
		gitlab.Endpoint.TokenURL, storage.PRIVATE_DATA.GL.ID, storage.PRIVATE_DATA.GL.SECRET, code, "https://hackoween.dev/oauth/gitlab")
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
	RedirectURL:  "https://hackoween.dev/oauth/google/callback",
	ClientID:     storage.PRIVATE_DATA.GG.ID,
	ClientSecret: storage.PRIVATE_DATA.GG.SECRET,
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
	error_log.CheckErr(err, error_log.Panic, "Failed Code Exchange")

	res, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	error_log.CheckErr(err, error_log.Panic,  "Failed GET on Google User Info")

	defer res.Body.Close()

	var user_info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&user_info); err != nil {
		error_log.CheckErr(err, error_log.Panic,  "Failed JSON Decode on Google User Info")
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

const AUTH_DATABASE = storage.DB
var AUTH_ID_SALT = storage.PRIVATE_DATA.OTHER.SALT

func hashAuthID(auth_id_raw string) string {
	return ToSHA(auth_id_raw + AUTH_ID_SALT)
}

func compareAuthID(auth_id_raw, auth_id string) bool {
	return hashAuthID(auth_id_raw) == auth_id
}

func addNewUser(auth_id_raw, name, username, pfp string, w http.ResponseWriter, r *http.Request) {
	auth_id := hashAuthID(auth_id_raw)
	session_key := generateSessionKey(auth_id)
	login_date := strings.Split(time.Now().String(), " ")[0]

	log.Printf("AUTH_ID: %x\n", auth_id)
	log.Println("Name:", name)
	log.Println("Username:", username)
	log.Println("Profile Pic:", pfp)
	log.Printf("Session Key: %x\n", session_key)
	log.Println("Login Date:", login_date)

	storage.InsertIntoTable(storage.AUTH_TABLE, "(auth_id, name, username, pfp, session_key, login_date)", auth_id, name, username, pfp, session_key, login_date)

	id_interface := storage.GetFromTable(storage.AUTH_TABLE, "id", "auth_id", auth_id)
	var id int
	if id_interface != nil {
		id = int(id_interface.(int64))
	} else {
		log.Fatal("FATAL | Critical Error in addNewUser: id is nil.")
	}

	anon_name := makeAnonName(id)
	log.Println("Anon Name:", anon_name)

	storage.UpdateTable(storage.AUTH_TABLE, "anon_name", anon_name, "id", id)

	storage.InsertIntoTable(storage.SETTINGS_TABLE, "id", id)

	storage.InsertIntoTable(storage.ALPHA_TABLE, "id", id)
	storage.InsertIntoTable(storage.HOW_2022_TABLE, "id", id)

	cookie.StoreCookie("session_key", session_key, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func makeAnonName(id int) string {
	return "Anon#"+strconv.Itoa(id)
}

func generateSessionKey(auth_id string) string {
	return ToSHA(auth_id + strconv.FormatInt(time.Now().Unix(), 10))
}

func CheckExistingUser(auth_id_raw string) bool {
	auth_id := hashAuthID(auth_id_raw)
	db, err := sql.Open("sqlite3", AUTH_DATABASE)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Auth Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE auth_id = ?)", storage.AUTH_TABLE))
	error_log.CheckErr(err, error_log.Fatal, "Failed Transaction Preparation")
	defer stmt.Close()

	var exists int
	stmt.QueryRow(auth_id).Scan(&exists)

	return exists == 1
}

// "auth_id" MUST be a valid user auth_id
func LoginUser(auth_id_raw string, w http.ResponseWriter, r *http.Request) {
	auth_id := hashAuthID(auth_id_raw)
	session_key := generateSessionKey(auth_id)
	login_date := strings.Split(time.Now().String(), " ")[0]

	db, err := sql.Open("sqlite3", AUTH_DATABASE)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Auth Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET session_key=?, login_date=? WHERE auth_id=?", storage.AUTH_TABLE))

	error_log.CheckErr(err, error_log.Fatal, "Failed Transaction Preparation")
	defer stmt.Close()

	_, err = stmt.Exec(session_key, login_date, auth_id)
	error_log.CheckErr(err, error_log.Fatal, "Failed Transaction Execution")

	cookie.StoreCookie("session_key", session_key, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func SignOutUser(w http.ResponseWriter, r *http.Request) {
	cookie.StoreCookie("session_key", nil, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	session_key_interface := cookie.GetCookie("session_key", r)
	var session_key string
	if session_key_interface != nil {
		session_key = session_key_interface.(string)
	} else {
		log.Panic("PANIC | session_key is nil in deletion")
		return
	}

	id := storage.GetFromTable_SessionKey(storage.AUTH_TABLE, session_key, "id")

	storage.DeleteFromTable(storage.CURR_EVENT_TABLE, "id", id)
	storage.DeleteFromTable(storage.SETTINGS_TABLE, "id", id)
	storage.DeleteFromTable(storage.AUTH_TABLE, "id", id)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
