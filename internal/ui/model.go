package ui

import (
	"fmt"
	"slugbug/internal/slugbug"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const maxWidth = 80

// Given a list of bus names, ask the user which they'd like to listen on
func ChooseServices(services []string) *huh.Group {
	return huh.NewGroup(
		huh.NewMultiSelect[string]().
			Key("name").
			Title("Choose your service").
			OptionsFunc(func() []huh.Option[string] {
				return huh.NewOptions(services...)
			}, &services),
	)
}

//////////////////////////////////////////////////////////////

type Model struct {
	form      *huh.Form
	hasDarkBg bool
	width     int
	styles    func(bool) *Styles
}

func NewModel() Model {

	// introduce user to the application
	introductionMsg := huh.NewNote().Title("Welcome to Slugbug.").
		Description("Gain visibility into DBus signals and methods.").
		Next(true).NextLabel("Connect to the DBus")

	// user must make some kind of action
	pickOperation := huh.NewSelect[string]().
		Options(huh.NewOptions(
			string(slugbug.Display),
			string(slugbug.Call),
			string(slugbug.Inspect),
			string(slugbug.Monitor),
		)...).
		Title("What would you like to do?").
		Key("operation")

	return Model{
		width:  maxWidth,
		styles: NewStyles,
		form: huh.NewForm(
			huh.NewGroup(introductionMsg),
			huh.NewGroup(pickOperation),
		).WithShowHelp(false).
			WithShowErrors(false).
			WithWidth(maxWidth),
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.hasDarkBg = msg.IsDark()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "esc", "q":
			return m, tea.Quit
		}
	}

	// pass any commands through to the huh form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	return m, cmd
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) View() tea.View {
	s := m.styles(m.hasDarkBg)

	switch m.form.State {
	case huh.StateCompleted:
		operation := m.form.Get("operation")
		var b strings.Builder

		fmt.Fprintf(&b, "You are choosing to execute %s\n", operation)
		return tea.NewView(s.Base.Padding(1, 2).Render(b.String()) + "\n\n")
	default:

		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := lipgloss.NewStyle().Margin(1, 0).Render(v)

		errors := m.form.Errors()
		body := lipgloss.JoinHorizontal(lipgloss.Left, form)
		header := m.appBoundaryView("Slugbug")
		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}

		return tea.NewView(s.Base.Render(header + "\n" + body + "\n\n" + footer))
	}

}

// convert form errors into a readable format
func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

// display a header boundary on the top and bottom sides of the form
func (m Model) appBoundaryView(text string) string {
	s := m.styles(m.hasDarkBg)
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		s.StatusBarText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(s.Indigo)),
	)
}

// display an error inside the top side of the form
func (m Model) appErrorBoundaryView(text string) string {
	s := m.styles(m.hasDarkBg)
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		s.ErrorText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(s.Red)),
	)
}
