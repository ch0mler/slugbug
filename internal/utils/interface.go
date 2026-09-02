package utils

import "charm.land/huh/v2"

// Given a list of bus names, ask the user which they'd like to listen on
func ChooseBus(busNames []string, ret *[]string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("name").
				Value(ret).
				Title("Choose your bus").
				OptionsFunc(func() []huh.Option[string] {
					return huh.NewOptions(busNames...)
				}, &busNames),
		),
	)
}
