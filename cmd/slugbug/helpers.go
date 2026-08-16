package main

import (
	"os/user"
)

func IsRootUser() bool {
	currentUser, err := user.Current()
	return (err == nil) && (currentUser.Username == "root")
}
