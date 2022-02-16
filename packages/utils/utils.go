package utils

import (
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/storage"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Returns a slice of all the files in the directory matching the extension
func GetFilesInDir(dir, ext string) (ret []string) {
	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext || ext == "" {
			ret = append(ret, filepath.Join(dir, file.Name()))
		}
	}
	return
}

// Returns a file if it exists
func GetFileInDir(dir, fn string) *os.File {
	path := filepath.Join(dir, fn)
	if FileExists(path) {
		file, _ := os.Open(path)
		return file
	}
	return nil
}

// Returns a file's contents if it exists
func GetFileContentsInDir(dir, fn string) string {
	path := filepath.Join(dir, fn)
	if FileExists(path) {
		file, _ := ioutil.ReadFile(path)
		return string(file)
	}
	return ""
}


func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetSessionKey(r *http.Request) string {
	key := cookie.GetCookie("session_key", r)
	if key != nil {
		return key.(string)
	}
	return ""
}

func CheckExistingSession(r *http.Request) bool {
	session_key := cookie.GetCookie("session_key", r)
	if session_key == nil {
		return false
	}

	login_date_interface := storage.GetFromTable(storage.AUTH_TABLE, "login_date", "session_key", session_key)
	var login_date_str string

	if login_date_interface != nil {
		login_date_str = login_date_interface.(string)
	} else {
		log.Printf("LOG | nil login date from session key: %v", session_key)
		return false
	}

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

	return !(since.Hours() > time.Hour.Hours()*24*18) // Session Expires in 18 Days
}

func GenUserTemplateData(r *http.Request) map[string]interface{} {
	m := make(map[string]interface{})
	session_key := GetSessionKey(r)
	if session_key != "" {
		row := storage.GetAllFromTable_SessionKey(storage.AUTH_TABLE, session_key)
		if row != nil {
			var id, timeout int
			var auth_id, name, username, anon_name, pfp, session_key_filler, login_date string
			row.Scan(&id, &auth_id, &name, &username, &anon_name, &pfp, &session_key_filler, &login_date, &timeout)

			m["Username"] = storage.GetUserName(session_key)
			m["PFP"] = pfp
			m["Login"] = CheckExistingSession(r)
			m["Alpha_Parts"] = storage.GetPartsCompleted(storage.Alpha, session_key)
		}
	}

	return m
}
