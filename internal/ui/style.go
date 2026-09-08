package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Styles struct {
	Base,
	ErrorText,
	StatusBar,
	StatusBarText,
	Help lipgloss.Style

	Yellow, Red, Indigo, Green, Purple color.Color
}

func NewStyles(hasDarkBg bool) *Styles {
	var (
		s         = Styles{}
		lightDark = lipgloss.LightDark(hasDarkBg)
	)

	s.Red = lightDark(lipgloss.Color("#FE5F86"), lipgloss.Color("#FE5F86"))
	s.Yellow = lightDark(lipgloss.Color("#BCDF30"), lipgloss.Color("#A4C522"))
	s.Green = lightDark(lipgloss.Color("#02BA84"), lipgloss.Color("#02BF87"))
	s.Indigo = lightDark(lipgloss.Color("#5A56E0"), lipgloss.Color("#7571F9"))
	s.Purple = lightDark(lipgloss.Color("#A550DF"), lipgloss.Color("#A550DF"))

	s.Base = lipgloss.NewStyle().
		Padding(1, 4, 0, 1)

	s.ErrorText = s.StatusBarText.Foreground(s.Red)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("#343433"), lipgloss.Color("#C1C6B2"))).
		Background(lightDark(lipgloss.Color("#D9DCCF"), lipgloss.Color("#353533")))

	s.StatusBarText = lipgloss.NewStyle().
		Foreground(s.Yellow).
		Bold(true).
		Padding(0, 1, 0, 2)

	s.Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	return &s
}
