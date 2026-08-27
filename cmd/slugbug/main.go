package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	unique := flag.Bool("unique", false, "Include unique connection names (e.g. :1.0, :1.11)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	flag.Parse()

	logger := initLogging(*debug)

	// connect to the appropriate bus and prepare to handle signals from it
	busConn := connectToBus(*system, *private, logger)
	// close the connection at the end of the run
	defer closeConnection(busConn, logger)

	names := listBusServices(busConn, *unique, logger)

	form := ChooseBus(names)
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	busNames := form.Get("name")
	fmt.Printf("You chose to monitor the following buses: %v\n", busNames)
}
