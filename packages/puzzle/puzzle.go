package puzzle

import (
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/storage"
	"hack-o-ween-site/packages/utils"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// Loops through and binds all submission URLs to the correct function
func BindURLs(url string, e storage.Event) {
	switch (e) {
	case storage.Alpha:
		for i := 1; i <= 3; i++ {
			http.HandleFunc(filepath.Join(url, "p" + strconv.Itoa(i), "submit"), GetUserSubmission)
			http.HandleFunc(filepath.Join(url, "p" + strconv.Itoa(i), "input"), SendUserInput)
		}
	case storage.HoW_2022:
		for i := 1; i <= 13; i++ {
			http.HandleFunc(filepath.Join(url, "d" + strconv.Itoa(i), "submit"), GetUserSubmission)
			http.HandleFunc(filepath.Join(url, "d" + strconv.Itoa(i), "input"), SendUserInput)
		}

	}
}

func GetUserSubmission(w http.ResponseWriter, r *http.Request) {
	session_key := utils.GetSessionKey(r)
	if session_key != "" {
		timeout := storage.GetFromTable_SessionKey(storage.AUTH_TABLE, session_key, "timeout").(int64)

		if !utils.HasTimePassed(timeout, 60) {
			m := utils.GenUserTemplateData(r, "alpha")
			templates := utils.GetFilesInDir("templates/", ".html")
			templates = append(templates[:1], templates...)
			templates[0] = filepath.Join("templates", "timeout.html")

			log.Println(strings.Trim(r.URL.Path, "/submit"))
			m["Puzzle_Page"] = "/" + strings.Trim(r.URL.Path, "/submit")
			m["Puzzle"] = 0

			tmpl := template.Must(template.ParseFiles(templates...))
			tmpl.Execute(w, m)
			return
		}

		// Parse the URL to get the current puzzle
		puzzle_slice := strings.Split(r.URL.Path, "/")
		if len(puzzle_slice) < 2 {
			return
		}

		puzzle_str := puzzle_slice[2]
		puzzle, err := strconv.Atoi(string(puzzle_str[1]))
		puzzle--
		if err != nil {
			log.Println("ERROR: Invalid Puzzle URL")
		}

		templates := utils.GetFilesInDir("templates/", ".html")
		templates = append(templates[:1], templates...)
		if CheckUserSubmission(storage.Alpha, puzzle, session_key, r.FormValue("answer")) {
			// Input is correct, nav to congrats page, set time completed at
			storage.IncPart(storage.Alpha, puzzle, session_key)
			templates[0] = filepath.Join("templates", "correct.html")
		} else {
			// Input is incorrect, nav to try again/etc page, set one minute timer
			storage.SetPuzzleTimeout(session_key)
			templates[0] = filepath.Join("templates", "incorrect.html")
		}
		m := utils.GenUserTemplateData(r, "alpha")
		log.Println(strings.Trim(r.URL.Path, "/submit"))
		m["Puzzle_Page"] = "/" + strings.Trim(r.URL.Path, "/submit")
		m["Puzzle"] = 0

		tmpl := template.Must(template.ParseFiles(templates...))
		tmpl.Execute(w, m)
	} else {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func CheckUserSubmission(e storage.Event, puzzle int, session_key, answer_raw string) bool {
	event_id := storage.GetUserEventID(session_key, e)

	dir := filepath.Join("storage", "puzzles", storage.EVENT_TO_STRING[e], strconv.Itoa(puzzle+1), "output")

	part := storage.GetPartsCompleted(e, session_key)[puzzle]
	if part < 2 {
		solution_full := utils.GetFileContentsInDir(dir, strconv.Itoa(event_id))
		solution_raw := strings.Split(solution_full, "\n")[part]

		solution, err1 := strconv.Atoi(solution_raw)
		answer, err2 := strconv.Atoi(answer_raw)
		if err1 == nil && err2 == nil {
			return solution == answer
		} else {
			return solution_raw == answer_raw
		}
	}
	return false
}

func SendUserInput(w http.ResponseWriter, r *http.Request) {
	// Parse the URL to get the current puzzle
	puzzle_slice := strings.Split(r.URL.Path, "/")
	if len(puzzle_slice) < 2 {
		return
	}

	puzzle_str := puzzle_slice[2]
	puzzle, err := strconv.Atoi(string(puzzle_str[1]))
	if err != nil {
		log.Println("ERROR: Invalid Puzzle URL")
	}

	key := cookie.GetCookie("session_key", r)
	session_key := ""
	if key != nil {
		session_key = key.(string)
		event_id := storage.GetUserEventID(session_key, storage.Alpha)
		w.Write([]byte(utils.GetFileContentsInDir(filepath.Join("storage", "puzzles", "alpha", strconv.Itoa(puzzle), "input"), strconv.Itoa(event_id))))
	}

}
