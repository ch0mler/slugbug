package main

import "charm.land/huh/v2"

func ChooseBus(busNames []string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("name").
				Title("Choose your bus").
				OptionsFunc(func() []huh.Option[string] {
					return huh.NewOptions(busNames...)
				}, &busNames),
		),
	)
}
