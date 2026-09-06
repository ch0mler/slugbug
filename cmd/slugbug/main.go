package main

import (
	"flag"
	"fmt"
	"os"
	"slugbug/internal/slugbug"
	"slugbug/internal/ui"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/list"
)

// Show the user a well-formatted list of service names according to their options
func displayServices(myName string, serviceNames []string) {
	svcList := list.New(serviceNames).
		Enumerator(list.Dash).
		EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99")).MarginRight(1)).
		ItemStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("ff")).MarginRight(1))

	lipgloss.Println(svcList)

	// Status Bar.
	var (
		// Detect the background color.
		hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
		lightDark = lipgloss.LightDark(hasDarkBG)

		statusBarStyle = lipgloss.NewStyle().
				Foreground(lightDark(lipgloss.Color("#343433"), lipgloss.Color("#C1C6B2"))).
				Background(lightDark(lipgloss.Color("#D9DCCF"), lipgloss.Color("#353533")))

		statusStyle = lipgloss.NewStyle().
				Inherit(statusBarStyle).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#A550DF")).
				Padding(0, 1).
				MarginRight(1)

		statusText = lipgloss.NewStyle().Inherit(statusBarStyle)
	)

	w := lipgloss.Width

	width := 96
	doc := strings.Builder{}
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)

	statusKey := statusStyle.Render("CONNECTED")
	statusVal := statusText.
		Width(width-w(statusKey)).
		Render("Name:", myName)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
	)

	doc.WriteString(statusBarStyle.Width(width).Render(bar))

	// Render the document.
	document := docStyle.Render(doc.String())
	lipgloss.Println(document)
}

func main() {
	unique := flag.Bool("unique", false, "Include unique connection names (e.g. :1.0, :1.11)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	system := flag.Bool("system", false, "Use system bus instead of session bus (requires elevated privileges)")
	private := flag.Bool("private", false, "Use a private connection to dbus")
	flag.Parse()

	// connect to the appropriate bus and prepare to handle signals from it
	sb := slugbug.NewSlugBug(*system, *private, *debug, *unique)
	// close the connection at the end of the run
	defer sb.CloseConnection()

	_, err := tea.NewProgram(ui.NewModel()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
	// userInterface := utils.New
	// form := utils.NewInterface(&operation)
	// for {
	// 	if err = form.Run(); err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	if operation == string(utils.Quit) {
	// 		break
	// 	} else if operation == string(utils.Display) {
	// 		allNames := sb.ListBusServices()
	// 		displayServices(sb.Name(), allNames)
	// 	} else if operation == string(utils.Inspect) {
	// 		fmt.Println("ERROR: Not implemented yet")
	// 	} else {
	// 		log.Fatalf("Unknown operation: %s\n", operation)
	// 	}
	// }
}
