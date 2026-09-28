package main

import (
	"flag"
	"fmt"
	"log"
	"slugbug/internal/slugbug"
)

func main() {
	unique := flag.Bool("unique", false, "Include unique connection names (e.g. :1.0, :1.11)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	flag.Parse()

	sb := slugbug.NewSlugBug(*system, *private, *debug, *unique)

	// connect to the appropriate bus and prepare to handle signals from it
	if err := sb.ConnectToBus(); err != nil {
		log.Fatal(err)
	}
	// close the connection at the end of the run
	defer sb.CloseConnection()

	services, err := sb.ListBusServices()
	if err != nil {
		log.Fatalf("Could not list bus services: %s", err)
	}
	for _, service := range services {
		fmt.Println(service)
	}
	/*
		program := tea.NewProgram(ui.NewModel(sb))
		if _, err := program.Run(); err != nil {
			log.Fatalf("Error running Slugbug TUI: %s", err.Error())
		}
	*/
}
