package main

import (
	"log/slog"
	"os"
	"os/user"
)

func initLogging(debug bool) *slog.Logger {
	logOpts := slog.HandlerOptions{Level: slog.LevelInfo}

	if debug {
		logOpts.Level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &logOpts))
}

func IsRootUser() bool {
	currentUser, err := user.Current()
	return (err == nil) && (currentUser.Username == "root")
}
