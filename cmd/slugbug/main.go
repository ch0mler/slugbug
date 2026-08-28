package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

func InspectBus(service string, conn *dbus.Conn) {
	node, err := introspect.Call(conn.Object(service, conn.BusObject().Path()))
	if err != nil {
		log.Fatal(err)
	}
	// nodeDetails := introspect.NewIntrospectable(node)
	// fmt.Println(nodeDetails.Introspect())
	for _, v := range node.Interfaces {
		fmt.Printf("Signals for %s\n", v.Name)
		for _, k := range v.Signals {
			fmt.Printf("%v\n", k)
		}
	}
}

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

	var (
		// all possible bus names, including unique connections
		allNames []string
		// only bus names the user is likely interested in
		busNames []string
	)

	allNames = listBusServices(busConn, *unique, logger)

	form := ChooseBus(allNames, &busNames)
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("You chose to monitor the following buses: %v\n", busNames)

	for _, v := range busNames {
		InspectBus(v, busConn)
	}
}
