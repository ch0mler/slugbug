package main

import (
	"flag"
	"fmt"
	"log"
	"slugbug/internal/slugbug"
	"slugbug/internal/ui"
)

func chooseOperation(sb *slugbug.Slugbug, operation string) {
	switch operation {
	case "list":
		services, err := sb.ListBusServices()
		if err != nil {
			log.Fatalf("Could not list bus services: %s", err)
		}
		for _, service := range services {
			fmt.Println(service)
		}
	case "inspect":
		fmt.Println("Inspect not implemented yet")
		service, objectPath := ui.InspectForm()
		if service == "" {
			log.Fatalf("No service specified for inspection")
		}
		result, err := sb.InspectService(service, objectPath)
		if err != nil {
			log.Fatalf("Could not inspect service %s: %s", service, err)
		}
		fmt.Print(result)
	case "invoke":
		fmt.Println("Invoke not implemented yet")
		// service := flag.Arg(1)
		// method := flag.Arg(2)
		// if service == "" || method == "" {
		// 	log.Fatalf("Service and method must be specified for invocation")
		// }
		// err := sb.InvokeMethod(service, method)
		// if err != nil {
		// 	log.Fatalf("Could not invoke method %s on service %s: %s", method, service, err)
		// }
	case "monitor":
		fmt.Println("Monitor not implemented yet")
		// service := flag.Arg(1)
		// signal := flag.Arg(2)
		// if service == "" || signal == "" {
		// 	log.Fatalf("Service and signal must be specified for monitoring")
		// }
		// err := sb.MonitorSignal(service, signal)
		// if err != nil {
		// 	log.Fatalf("Could not monitor service %s: %s", service, err)
		// }
	default:
		if operation == "" {
			fmt.Printf("No operation specified\n")
		} else {
			fmt.Printf("Unknown operation: %s\n", operation)
		}

		log.Fatalf("Acceptable operations are: list, inspect, invoke, monitor")
	}
}

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

	operation := flag.Arg(0)

	chooseOperation(sb, operation)
	/*
		program := tea.NewProgram(ui.NewModel(sb))
		if _, err := program.Run(); err != nil {
			log.Fatalf("Error running Slugbug TUI: %s", err.Error())
		}
	*/
}
