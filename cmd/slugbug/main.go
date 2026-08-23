package main

import (
	"flag"
	"fmt"

	"github.com/godbus/dbus/v5"
)

func listBus(conn *dbus.Conn) {
	fmt.Printf("Bus: %v\n", conn.BusObject())
	fmt.Printf("Names: %v\n", conn.Names())
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	watch := flag.Bool("watch", false, "Monitor dbus for signals")
	flag.Parse()

	logger := initLogging(*debug)

	// connect to the appropriate bus and prepare to handle signals from it
	busConn := connectToBus(*system, *private, logger)

	if *watch {
		enableWatch(busConn, logger)
	} else {
		listBus(busConn)
	}
}
