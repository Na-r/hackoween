package storage

import (
	"io/fs"
	"log"
	"os"
	"sync"
)

const STORAGE_DIR = "storage/"

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
