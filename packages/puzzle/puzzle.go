package puzzle

import (
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/storage"
	"hack-o-ween-site/packages/utils"
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
		}
	case storage.HoW_2022:
		for i := 1; i <= 13; i++ {
			http.HandleFunc(filepath.Join(url, "d" + strconv.Itoa(i), "submit"), GetUserSubmission)
		}

	}
}

func GetUserSubmission(w http.ResponseWriter, r *http.Request) {
	// Parse the input to get the current puzzle
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
		if CheckUserSubmission(storage.Alpha, puzzle, session_key, r.FormValue("answer")) {
			// Input is correct, nav to congrats page
		} else {
			// Input is incorrect, nav to try again/etc page, set one minute timer
		}
	}

	// Temporary redirect while other pages are in development
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func CheckUserSubmission(e storage.Event, p int, session_key, answer_raw string) bool {
	event_id := storage.GetUserEventID(session_key, e)

	dir := "solutions/"
	switch (e) {
	case storage.Alpha:
		dir += "alpha/"
	case storage.HoW_2022:
		dir += "HoW2022/"
	}

	solution_raw := utils.GetFileContentsInDir(dir, strconv.Itoa(event_id))

	solution, err1 := strconv.Atoi(solution_raw)
	answer, err2 := strconv.Atoi(answer_raw)
	if err1 == nil && err2 == nil {
		return solution == answer
	} else {
		return solution_raw == answer_raw
	}
}
