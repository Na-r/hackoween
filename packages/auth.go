package packages

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	_ "golang.org/x/oauth2/google"
)

const clientID_github = "e3608b8becc8be75c377"
const clientSecret_github = "edf9f8aea66771b6cd9077325c4c52799cb98710"

const clientID_gitlab = "09e978b08b2056b0e7d98fd8af90a55f14038189473c162d82fa97e4bfe1b608"
const clientSecret_gitlab = "c0022910a61977e536e6ad688723db34c538d2de549e99054db8d675c008b65f"

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

	fmt.Println("Login Successful. Token is", t.AccessToken)

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

	id_f := user_info["id"].(float64)
	id := fmt.Sprintf("%.0f", id_f) + "GH"
	addNewUser(id, user_info["name"].(string),
		user_info["login"].(string), user_info["avatar_url"].(string))

	http.Redirect(w, r, "/", http.StatusFound)
	return
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
		log.Println("Error Retrieving Gitlab User Info")
		return
	} else if _, ok := user_info["name"]; !ok {
		log.Println("Error Retrieving Gitlab Access Info")
		return
	} else if _, ok := user_info["username"]; !ok {
		log.Println("Error Retrieving Gitlab Access Info")
		return
	} else if _, ok := user_info["avatar_url"]; !ok {
		log.Println("Error Retrieving Gitlab Access Info")
		return
	}

	id_f := user_info["id"].(float64)
	id := fmt.Sprintf("%.0f", id_f) + "GL"
	addNewUser(id, user_info["name"].(string),
		user_info["username"].(string), user_info["avatar_url"].(string))

	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func addNewUser(id, name, username, pfp string) {
	/* First generate the required parameters:
	Session Key (ID + UNIX Time, Hashed)
	Last Login (Current Date as YYYY-MM-DD)
	*/

	curr_time := time.Now().Unix()

	session_key := ToSHA(id + strconv.FormatInt(curr_time, 10))
	login_date := strings.Split(time.Now().String(), " ")[0]

	log.Println("ID:", id)
	log.Println("Name:", name)
	log.Println("Username:", username)
	log.Println("Profile Pic:", pfp)
	log.Printf("Session Key: %x\n", session_key)
	log.Println("Login Date:", login_date)

	return

	/*
		db, err := sql.Open("sqlite3", "./foo.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		stmt, err := tx.Prepare("insert into bar(auth_token, session_key, last_login) values(?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()
		for i := 0; i < 100; i++ {
			_, err = stmt.Exec(i, fmt.Sprintf("こんにちわ世界%03d", i))
			if err != nil {
				log.Fatal(err)
			}
		}
		tx.Commit()

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
}
