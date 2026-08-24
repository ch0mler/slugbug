package main

import (
	"flag"
	"fmt"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	watch := flag.Bool("watch", false, "Monitor dbus for signals")
	list := flag.Bool("list", false, "List D-Bus services")
	flag.Parse()

	logger := initLogging(*debug)

	// connect to the appropriate bus and prepare to handle signals from it
	busConn := connectToBus(*system, *private, logger)
	// close the connection at the end of the run
	defer closeConnection(busConn, logger)

	if *watch {
		enableWatch(busConn, logger)
	} else if *list {
		listBusServices(busConn, logger)
	} else {
		fmt.Printf("No action specified. Exiting\n")
	}
}
