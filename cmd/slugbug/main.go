package main

import (
	"flag"
	"fmt"
	"log"
	"slugbug/internal/slugbug"
	"slugbug/internal/utils"
)

func main() {
	unique := flag.Bool("unique", false, "Include unique connection names (e.g. :1.0, :1.11)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	flag.Parse()

	// connect to the appropriate bus and prepare to handle signals from it
	sb := slugbug.NewSlugBug(*system, *private, *debug)
	// close the connection at the end of the run
	defer sb.CloseConnection()

	var (
		// all possible bus names, including unique connections
		allNames []string
		// only bus names the user is likely interested in
		busNames []string
	)

	allNames = sb.ListBusServices(*unique)

	form := utils.ChooseBus(allNames, &busNames)
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("You chose to monitor the following buses: %v\n", busNames)

	for _, v := range busNames {
		sb.InspectBus(v)
	}
}
