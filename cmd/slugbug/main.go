package main

import (
	"flag"
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
	sb.ConnectToBus()
	// close the connection at the end of the run
	defer sb.CloseConnection()

	/*
	   // ask user which services they want to look at
	   	for _, v := range sb.ListBusServices() {
	   		fmt.Println(v)
	   	}

	   // display traits of a particular service
	   sb.InspectService("org.freedesktop.DBus")
	*/
}
