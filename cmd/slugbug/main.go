package main

import (
	"flag"
	"log"
	"os"
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

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	watch := flag.Bool("watch", false, "Monitor dbus for signals")
	flag.Parse()

	initLogging(*debug)

	// connect to the appropriate bus and prepare to handle signals from it
	busConn := connectToBus(*system)

	if *watch {
		enableWatch(busConn)
	}
}
