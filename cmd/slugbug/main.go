package main

import (
	"flag"
	"log"
	"slugbug/internal/slugbug"
	"slugbug/internal/ui"

	tea "charm.land/bubbletea/v2"
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

	program := tea.NewProgram(ui.NewModel(sb))
	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running Slugbug TUI: %s", err.Error())
	}
}
