package logger

import (
	"log"
	"os"
)

var (
	Info          = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error         = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
)
