package storage

import (
	"database/sql"
	"fmt"
	"hack-o-ween-site/packages/utils"
	"io/fs"
	"log"
	"os"
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

type ThemeType uint8
const (
	Dark ThemeType = iota
	Light
)

type NameType uint8
const (
    Username NameType = iota
    RealName
	Anonymous
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

func init() {
	_, err := os.OpenFile(STORAGE_DIR+UID.file, os.O_RDONLY|os.O_CREATE, fs.ModePerm)
	if err != nil {
		log.Fatalf("FATAL | Error Opening %s: %s\n", UID.file, err.Error())
	}
}

func InsertIntoTable(table_name, cols string, args ...interface{}) {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
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
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Insertion Transaction Preparation for Table %s, Columns %s", table_name, cols))
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Insertion Transaction Execution for Table %s, Columns %s", table_name, cols))
}

func UpdateTable(table_name, arg_name string, arg_val interface{}, where_name string, where_val interface{}) {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", table_name, arg_name, where_name))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Update Transaction Preparation for Table %s, Args %s", table_name, arg_name))
	defer stmt.Close()

	_, err = stmt.Exec(arg_val, where_val)
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Update Transaction Execution for Table %s, Args %s", table_name, arg_name))
}

func UpdateUserSettings(session_key string, name, theme int) {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE session_key=?", AUTH_TABLE))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(session_key).Scan(&id)

	if id == 0 {
		return
	}

	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %s SET name_type=?,theme_type=? WHERE id=?", SETTINGS_TABLE))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Update Transaction Preparation for Table %s, Args %s", SETTINGS_TABLE, "name_type, theme_type"))
	defer stmt.Close()

	_, err = stmt.Exec(name, theme, id)
	utils.CheckErr(err, utils.Log, fmt.Sprintf("Failed Update Transaction Execution for Table %s, Args %s", SETTINGS_TABLE, "name_type, theme_type"))
}

func GetFromTable(table_name, return_name, where_name string, where_val interface{}) interface{} {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("SELECT %s FROM %s WHERE %s=?", return_name, table_name, where_name))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", table_name, return_name))
	defer stmt.Close()

	var ret interface{}
	stmt.QueryRow(where_val).Scan(&ret)
	return ret
}

func GetFromTable_AuthID(table_name, auth_id, return_name string) interface{} {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE auth_id=?", AUTH_TABLE))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(auth_id).Scan(&id)

	stmt, err := db.Prepare(fmt.Sprintf("SELECT %s FROM %s WHERE id=?", return_name, table_name))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", table_name, return_name))
	defer stmt.Close()

	var ret interface{}
	stmt.QueryRow(id).Scan(&ret)
	return ret
}

func GetFromTable_SessionKey(table_name, session_key, return_name string) interface{} {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE session_key=?", AUTH_TABLE))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(session_key).Scan(&id)

	if id == 0 {
		return nil
	}

	stmt, err := db.Prepare(fmt.Sprintf("SELECT %s FROM %s WHERE id=?", return_name, table_name))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", table_name, return_name))
	defer stmt.Close()

	var ret interface{}
	stmt.QueryRow(id).Scan(&ret)
	return ret
}

func GetAllFromTable_SessionKey(table_name, session_key string) *sql.Row {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt_id, err := db.Prepare(fmt.Sprintf("SELECT id FROM %s WHERE session_key=?", AUTH_TABLE))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, Return Name %s", AUTH_TABLE, "id"))
	defer stmt_id.Close()

	var id int
	stmt_id.QueryRow(session_key).Scan(&id)

	if id == 0 {
		return nil
	}

	stmt, err := db.Prepare(fmt.Sprintf("SELECT * FROM %s WHERE id=?", table_name))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Get Transaction Preparation for Table %s, *", table_name))
	defer stmt.Close()

	return stmt.QueryRow(id)
}

func DeleteFromTable(table_name, where_name string, where_val interface{}) {
	db, err := sql.Open("sqlite3", DB)
	utils.CheckErr(err, utils.Fatal, "Failed Opening Database")
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %s WHERE %s=?", table_name, where_name))
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Deletion Transaction Preparation for Table %s, Where %s", table_name, where_name))
	defer stmt.Close()

	_, err = stmt.Exec(where_val)
	utils.CheckErr(err, utils.Fatal, fmt.Sprintf("Failed Deletion Transaction Update for Table %s, Where %s", table_name, where_name))
}
