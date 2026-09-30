package ui

import (
	"slugbug/internal/slugbug"

	"charm.land/huh/v2"
)

// ////////////////////////////////////////////////////////////
// Temporary forms while functionality is being implemented //
// ////////////////////////////////////////////////////////////

func InspectForm() (string, string) {
	var (
		service    string
		objectPath string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter the service name to inspect").
				Placeholder("org.freedesktop.systemd1").
				CharLimit(1000).
				Value(&service),
			huh.NewInput().
				Title("Enter the object path to inspect").
				Placeholder("/org/freedesktop/systemd1").
				CharLimit(1000).
				Value(&objectPath),
		),
	)
	// Note: no error checking here because it's only development
	// and we need some motivation to implement a real UI
	form.Run()
	return service, objectPath
}

////////////////////////////////////////////////////////////
// End of temporary forms
////////////////////////////////////////////////////////////

func operationOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(slugbug.AvailableOperations))
	for _, operation := range slugbug.AvailableOperations {
		options = append(options, huh.Option[string]{Key: operation.String(), Value: operation.String()})
	}
	return options
}

func IntroductionForm(ret *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Welcome to Slugbug.").
				Description("Gain visibility into DBus signals and methods.").
				Next(true).NextLabel("Connect to the DBus"),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(operationOptions()...).
				Title("What would you like to do?").
				Key("operation").
				Value(ret),
		),
	).WithShowHelp(true)
}

func OperationForm(ret *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(operationOptions()...).
				Title("What would you like to do?").
				Key("operation").
				Value(ret),
		),
	).WithShowHelp(true)
}

// Given a list of bus names, ask the user which they'd like to listen on
func ChooseServicesForm(services []string, ret *[]string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("services").
				Value(ret).
				Title("Choose your services").
				OptionsFunc(func() []huh.Option[string] {
					return huh.NewOptions(services...)
				}, &services),
		),
	).WithShowHelp(true)
}

func ChooseServiceForm(services []string, ret *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("service").
				Value(ret).
				Title("Choose your service").
				OptionsFunc(func() []huh.Option[string] {
					return huh.NewOptions(services...)
				}, &services),
		),
	).WithShowHelp(true)
}
