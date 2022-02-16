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
		if CheckUserSubmission(storage.Alpha, puzzle, session_key, r.FormValue("answer")) {
			// Input is correct, nav to congrats page, set time completed at
		} else {
			// Input is incorrect, nav to try again/etc page, set one minute timer
		}
	}

	// Temporary redirect while other pages are in development
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func CheckUserSubmission(e storage.Event, p int, session_key, answer_raw string) bool {
	event_id := storage.GetUserEventID(session_key, e)

	dir := filepath.Join("storage", "puzzles", storage.EVENT_TO_STRING[e], strconv.Itoa(p), "output")

	solution_raw := utils.GetFileContentsInDir(dir, strconv.Itoa(event_id))

	solution, err1 := strconv.Atoi(solution_raw)
	answer, err2 := strconv.Atoi(answer_raw)
	if err1 == nil && err2 == nil {
		return solution == answer
	} else {
		return solution_raw == answer_raw
	}
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

func UpdatePartsCompleted(e storage.Event, puzzle, status int, session_key string) {
	str := storage.GetFromTable_SessionKey(storage.EVENT_TO_STRING[e], session_key, "parts").(string)
	if puzzle >= len(str) {
		log.Panic("PANIC | Trying to access nonexistent puzzle in puzzles.UpdatePartsCompleted")
		return
	}

	spl := strings.Split(str, "")
	spl[puzzle] = strconv.Itoa(status)
	new_str := strings.Join(spl, "")

	storage.UpdateTable(storage.EVENT_TO_STRING[e], "parts", new_str, "session_key", session_key)
}

func IncPart(e storage.Event, puzzle int, session_key string) {
	str := storage.GetFromTable_SessionKey(storage.EVENT_TO_STRING[e], session_key, "parts").(string)
	if puzzle >= len(str) {
		log.Panic("PANIC | Trying to increment nonexistent puzzle in puzzles.IncPart")
		return
	}

	spl := strings.Split(str, "")
	part, _ := strconv.Atoi(spl[puzzle])
	spl[puzzle] = strconv.Itoa(part+1)
	new_str := strings.Join(spl, "")

	storage.UpdateTable(storage.EVENT_TO_STRING[e], "parts", new_str, "session_key", session_key)
}

func GetPartsCompleted(e storage.Event, session_key string) []int {
	str := storage.GetFromTable_SessionKey(storage.EVENT_TO_STRING[e], session_key, "parts").(string)
	arr := []int{}
	for _, r := range str {
		i, _ := strconv.Atoi(string(r))
		arr = append(arr, i)
	}
	return arr
}
