package ui

import (
	"slugbug/internal/slugbug"

	"charm.land/huh/v2"
)

func IntroductionForm(ret *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Welcome to Slugbug.").
				Description("Gain visibility into DBus signals and methods.").
				Next(true).NextLabel("Connect to the DBus"),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(huh.NewOptions(
					string(slugbug.Display),
					string(slugbug.Call),
					string(slugbug.Inspect),
					string(slugbug.Monitor),
				)...).
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
