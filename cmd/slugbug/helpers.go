package main

import (
	"log"
	"os"
	"os/user"
)

// initialize logging functionality
func initLogging(debug bool) {
	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stderr)
	}
}

func IsRootUser() bool {
	currentUser, err := user.Current()
	return (err == nil) && (currentUser.Username == "root")
}
