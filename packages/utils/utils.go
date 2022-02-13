package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type ErrorLevel uint8

const (
	Log ErrorLevel = iota
	Panic
	Fatal
)

func CheckErr(err error, error_type ErrorLevel, desc string) {
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

func GetFilesInDir(dir, ext string) (ret []string){
	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext || ext == "" {
			ret = append(ret, filepath.Join(dir, file.Name()))
		}
	}
	return
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
