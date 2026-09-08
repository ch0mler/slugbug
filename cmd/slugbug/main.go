package main

import (
	"flag"
	"fmt"
	"log"
	"slugbug/internal/slugbug"
	"slugbug/internal/ui"

	"charm.land/huh/v2"
)

var (
	activeForm *huh.Form
	operation  string
)

func main() {
	unique := flag.Bool("unique", false, "Include unique connection names (e.g. :1.0, :1.11)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	flag.Parse()

	sb := slugbug.NewSlugBug(*system, *private, *debug, *unique)

	activeForm := ui.IntroductionForm(&operation)
	if err := activeForm.Run(); err != nil {
		log.Fatalf("Error running IntroductionForm: %s", err.Error())
	}
	operation := activeForm.Get("operation")

	// connect to the appropriate bus and prepare to handle signals from it
	sb.ConnectToBus()
	// close the connection at the end of the run
	defer sb.CloseConnection()

	fmt.Printf("Found %d services\n", len(sb.Services()))

	switch operation {
	case string(slugbug.Display):
		// ask user which services they want to look at
		for _, v := range sb.Services() {
			fmt.Println(v)
		}
	case string(slugbug.Monitor):
		var monitorSvc string
		activeForm := ui.ChooseServiceForm(sb.Services(), &monitorSvc)
		if err := activeForm.Run(); err != nil {
			log.Fatalf("Error running ChooseServiceForm: %s", err.Error())
		}
		fmt.Printf("You chose the following service: %v\n", monitorSvc)
		// METHOD NOT IMPLEMENTED
		fmt.Println("Method not implemented")
	case string(slugbug.Inspect):
		var inspectSvc string
		activeForm := ui.ChooseServiceForm(sb.Services(), &inspectSvc)
		if err := activeForm.Run(); err != nil {
			log.Fatalf("Error running ChooseServiceForm: %s", err.Error())
		}
		fmt.Printf("activeForm.service returned: %s\n", activeForm.GetString("service"))
		fmt.Printf("You chose the following service: %v\n", inspectSvc)
		sb.InspectService(inspectSvc)
	case string(slugbug.Call):
		var callSvc string
		activeForm := ui.ChooseServiceForm(sb.Services(), &callSvc)
		if err := activeForm.Run(); err != nil {
			log.Fatalf("Error running ChooseServiceForm: %s", err.Error())
		}
		fmt.Printf("You chose the following service: %v\n", callSvc)
		// METHOD NOT IMPLEMENTED
		fmt.Println("Method not implemented")
	default:
		log.Fatalf("Unknown operation: %v\n", operation)
	}
}
