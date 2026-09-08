package helpers

import (
	"log"
	"log/slog"
	"os"
	"os/user"
)

const LOG_FILENAME = "slugbug.log"

func InitLogging(debug bool) *slog.Logger {
	logOpts := slog.HandlerOptions{Level: slog.LevelInfo}

	if debug {
		logOpts.Level = slog.LevelDebug
		logFile, err := os.OpenFile(LOG_FILENAME, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Unable to open log file for writing: %s\n", LOG_FILENAME)
		}

		return slog.New(slog.NewTextHandler(logFile, &logOpts))
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &logOpts))
}

func IsRootUser() bool {
	currentUser, err := user.Current()
	return (err == nil) && (currentUser.Username == "root")
}

func LogFatal(message string, err error, logger *slog.Logger) {
	logger.Error(message, slog.String("error", err.Error()))
	os.Exit(1)
}
