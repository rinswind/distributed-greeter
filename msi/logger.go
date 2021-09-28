package msi

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lshortfile)
)
