package storage

import (
	"database/sql"
	"fmt"
	"hack-o-ween-site/packages/error_log"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const STORAGE_DIR = "storage/"
const DB = STORAGE_DIR+"HoW.db"

const AUTH_TABLE = "Auth"
const SETTINGS_TABLE = "Settings"
const ALPHA_TABLE = "Alpha"
const HOW_2022_TABLE = "HoW_2022"
const CURR_EVENT_TABLE = ALPHA_TABLE

type Event uint8
const (
	Alpha Event = 0
	HoW_2022 Event = 1
)

type ThemeType uint8
const (
	Dark ThemeType = 0
	Light ThemeType = 1
)

type NameType uint8
const (
    Username NameType = 0
    RealName NameType = 1
	Anonymous NameType = 2
)

type UserSettings struct {
	Theme ThemeType
	NameSetting NameType
}

type DaysInfo struct {
	Completed bool
	PuzzlesSolved int
	DateCompleted time.Time
}

type ExtraInfo struct {
	Completed bool
}

type UserInfo struct {
	Settings UserSettings
	Days	[]DaysInfo
}

type UID_Struct struct {
	mu      sync.Mutex
	curr_id uint
	file    string `default:"test"`
}

var UID UID_Struct = UID_Struct{
	curr_id: 0,
	file:    "id.cdb",
}

var EVENT_TO_STRING map[Event]string

func init() {
	EVENT_TO_STRING = make(map[Event]string)
	EVENT_TO_STRING[Alpha] = strings.ToLower(ALPHA_TABLE)
	EVENT_TO_STRING[HoW_2022] = strings.ToLower(HOW_2022_TABLE)
}

func InsertIntoTable(table_name, cols string, args ...interface{}) {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	if rune(cols[0]) != rune('(') {
		cols = "(" + cols + ")"
	}

	vals := ""
	for i := 0; i < len(args)-1; i++ {
		vals += "?, "
	}
	vals = "(" + vals + "?)"

	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s %s VALUES%s", table_name, cols, vals))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Insertion Transaction Preparation for Table %s, Columns %s", table_name, cols))
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Insertion Transaction Execution for Table %s, Columns %s", table_name, cols))
}

func UpdateTable(table_name, arg_name string, arg_val interface{}, where_name string, where_val interface{}) {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", table_name, arg_name, where_name))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Update Transaction Preparation for Table %s, Args %s", table_name, arg_name))
	defer stmt.Close()

	_, err = stmt.Exec(arg_val, where_val)
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Update Transaction Execution for Table %s, Args %s", table_name, arg_name))
}

func UpdateUserSettings(session_key string, name, theme int) {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE session_key=?", AUTH_TABLE))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(session_key).Scan(&id)

	if id == 0 {
		return
	}

	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET name_type=?,theme_type=? WHERE id=?", SETTINGS_TABLE))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Update Transaction Preparation for Table %s, Args %s", SETTINGS_TABLE, "name_type, theme_type"))
	defer stmt.Close()

	_, err = stmt.Exec(name, theme, id)
	error_log.CheckErr(err, error_log.Log, fmt.Sprintf("Failed Update Transaction Execution for Table %s, Args %s", SETTINGS_TABLE, "name_type, theme_type"))
}

func GetFromTable(table_name, return_name, where_name string, where_val interface{}) interface{} {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("SELECT %s FROM %s WHERE %s=?", return_name, table_name, where_name))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", table_name, return_name))
	defer stmt.Close()

	var ret interface{}
	stmt.QueryRow(where_val).Scan(&ret)
	return ret
}

func GetFromTable_AuthID(table_name, auth_id, return_name string) interface{} {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE auth_id=?", AUTH_TABLE))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(auth_id).Scan(&id)

	stmt, err := db.Prepare(fmt.Sprintf("SELECT %s FROM %s WHERE id=?", return_name, table_name))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", table_name, return_name))
	defer stmt.Close()

	var ret interface{}
	stmt.QueryRow(id).Scan(&ret)
	return ret
}

func GetFromTable_SessionKey(table_name, session_key, return_name string) interface{} {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE session_key=?", AUTH_TABLE))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(session_key).Scan(&id)

	if id == 0 {
		return nil
	}

	stmt, err := db.Prepare(fmt.Sprintf("SELECT %s FROM %s WHERE id=?", return_name, table_name))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", table_name, return_name))
	defer stmt.Close()

	var ret interface{}
	stmt.QueryRow(id).Scan(&ret)
	return ret
}

func GetAllFromTable_SessionKey(table_name, session_key string) *sql.Row {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE session_key=?", AUTH_TABLE))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(session_key).Scan(&id)

	if id == 0 {
		return nil
	}

	stmt, err := db.Prepare(fmt.Sprintf("SELECT * FROM %s WHERE id=?", table_name))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, *", table_name))
	defer stmt.Close()

	return stmt.QueryRow(id)
}

func DeleteFromTable(table_name, where_name string, where_val interface{}) {
	db, err := sql.Open("sqlite3", DB)
	error_log.CheckErr(err, error_log.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE %s=?", table_name, where_name))
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Deletion Transaction Preparation for Table %s, Where %s", table_name, where_name))
	defer stmt.Close()

	_, err = stmt.Exec(where_val)
	error_log.CheckErr(err, error_log.Fatal, fmt.Sprintf("Failed Deletion Transaction Update for Table %s, Where %s", table_name, where_name))
}

func GetUserName(session_key string) string {
	row := GetAllFromTable_SessionKey(AUTH_TABLE, session_key)
	var id, timeout int
	var auth_id, name, username, anon_name, pfp, session_key_filler, login_date string
	row.Scan(&id, &auth_id, &name, &username, &anon_name, &pfp, &session_key_filler, &login_date, &timeout)

	name_setting_intf := GetFromTable_SessionKey(SETTINGS_TABLE, session_key, "name_type")
	var name_setting NameType
	if name_setting_intf != nil {
		name_setting = NameType(name_setting_intf.(int64))
	}

	switch (name_setting) {
	case Username:
		if username != "" {
			return username
		} else if name != "" {
			return name
		} else {
			return anon_name
		}

	case RealName:
		if name != "" {
			return name
		} else {
			return anon_name
		}

	case Anonymous:
		return anon_name
	}

	return anon_name
}

func GetUserEventID(session_key string, e Event) int {
	var table string
	switch (e) {
	case Alpha:
		table = ALPHA_TABLE
	case HoW_2022:
		table = HOW_2022_TABLE
	}

	id_intf := GetFromTable_SessionKey(table, session_key, "event_id")

	var event_id int
	if id_intf != nil {
		event_id = int(id_intf.(int64))
	}

	if event_id == -1 {
		event_id = rand.Intn(1000)
		id_intf = GetFromTable_SessionKey(table, session_key, "id")
		var id int
		if id_intf != nil {
			id = int(id_intf.(int64))
		} else {
			return -100
		}
		UpdateTable(table, "event_id", event_id, "id", id)
	}

	return event_id
}

func UpdatePartsCompleted(e Event, puzzle, status int, session_key string) {
	str := GetFromTable_SessionKey(EVENT_TO_STRING[e], session_key, "parts").(string)
	if puzzle >= len(str) {
		log.Panic("PANIC | Trying to access nonexistent puzzle in puzzles.UpdatePartsCompleted")
		return
	}

	spl := strings.Split(str, "")
	spl[puzzle] = strconv.Itoa(status)
	new_str := strings.Join(spl, "")

	UpdateTable(EVENT_TO_STRING[e], "parts", new_str, "session_key", session_key)
}

func IncPart(e Event, puzzle int, session_key string) {
	str := GetFromTable_SessionKey(EVENT_TO_STRING[e], session_key, "parts").(string)
	if puzzle >= len(str) {
		log.Panic("PANIC | Trying to increment nonexistent puzzle in puzzles.IncPart")
		return
	}

	spl := strings.Split(str, "")
	part, _ := strconv.Atoi(spl[puzzle])
	spl[puzzle] = strconv.Itoa(part+1)
	new_str := strings.Join(spl, "")

	id := GetFromTable_SessionKey(AUTH_TABLE, session_key, "id")

	if id == nil {
		return
	}

	UpdateTable(EVENT_TO_STRING[e], "parts", new_str, "id", id)
}

func GetPartsCompleted(e Event, session_key string) []int {
	str := GetFromTable_SessionKey(EVENT_TO_STRING[e], session_key, "parts").(string)
	arr := []int{}
	for _, r := range str {
		i, _ := strconv.Atoi(string(r))
		arr = append(arr, i)
	}
	return arr
}
