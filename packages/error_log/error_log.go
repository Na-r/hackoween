package error_log

import "log"

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
