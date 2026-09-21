package ui

import (
	"fmt"
	"slugbug/internal/slugbug"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const maxWidth = 80

///////////////////////////////////////
//		Basic interface methods		 //
///////////////////////////////////////

type Model struct {
	slugbug    *slugbug.Slugbug
	form       *huh.Form
	screen     screen
	operation  string
	service    string
	result     string
	err        error
	resultView viewport.Model
	hasDarkBg  bool
	width      int
	styles     func(bool) *Styles
}

type screen int

const (
	operationScreen screen = iota
	serviceScreen
	displayScreen
	resultScreen
)

type operationResultMsg struct {
	text string
	err  error
}

type serviceListMsg struct {
	services []string
	err      error
}

func NewModel(sb *slugbug.Slugbug) Model {
	m := Model{
		slugbug: sb,
		screen:  operationScreen,
		width:   maxWidth,
		styles:  NewStyles,
	}
	m.form = IntroductionForm(&m.operation)
	m.resultView = viewport.New(
		viewport.WithWidth(maxWidth),
		viewport.WithHeight(24),
	)
	m.resultView.SoftWrap = true
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.hasDarkBg = msg.IsDark()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "q":
			if m.screen == displayScreen {
				break
			}
			return m, tea.Quit
		case "esc":
			if m.screen != operationScreen {
				return m.toOperationScreen()
			}
		case "enter", " ", "space":
			if m.screen == resultScreen {
				return m.toOperationScreen()
			}
		}
	}

	if services, ok := msg.(serviceListMsg); ok {
		if services.err != nil {
			m.screen = resultScreen
			m.result = ""
			m.err = services.err
			m.resetResultView()
			return m, nil
		}
		m.screen = serviceScreen
		m.form = ChooseServiceForm(services.services, &m.service)
		return m, m.form.Init()
	}

	if result, ok := msg.(operationResultMsg); ok {
		m.screen = resultScreen
		m.result = result.text
		m.err = result.err
		m.resetResultView()
		return m, nil
	}

	if services, ok := msg.(displayServicesMsg); ok {
		if services.err != nil {
			m.screen = resultScreen
			m.result = ""
			m.err = services.err
			m.resetResultView()
			return m, nil
		}
		m.screen = displayScreen
		m.form = ChooseServiceForm(services.services, &m.service)
		return m, nil
	}

	if m.screen == resultScreen {
		var cmd tea.Cmd
		m.resultView, cmd = m.resultView.Update(msg)
		return m, cmd
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		if m.screen == operationScreen {
			m.operation = m.form.GetString("operation")
			return m.dispatchOperation()
		}

		m.service = m.form.GetString("service")
		if m.screen == displayScreen {
			return m.toOperationScreen()
		}
		m.screen = resultScreen
		m.result = ""
		m.err = nil
		m.resetResultView()
		return m, inspectCmd(m.slugbug, m.service)
	}

	return m, cmd
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) View() tea.View {
	s := m.styles(m.hasDarkBg)

	if m.screen == displayScreen {
		header := m.appBoundaryView("Slugbug")
		footer := m.footerView()
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		body := lipgloss.NewStyle().Margin(1, 0).Render(v)
		return tea.NewView(s.Base.Render(header + "\n" + body + "\n\n" + footer))
	}

	if m.screen == resultScreen {
		header := m.appBoundaryView("Slugbug")
		if m.err != nil {
			header = m.appErrorBoundaryView(m.err.Error())
		}
		body := m.resultView.View() + "\n\nPress Enter or Space to return to the operation menu."
		footer := m.footerView()
		return tea.NewView(s.Base.Render(header + "\n" + body + "\n\n" + footer))
	}

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := lipgloss.NewStyle().Margin(1, 0).Render(v)

	errors := m.form.Errors()
	body := lipgloss.JoinHorizontal(lipgloss.Left, form)
	header := m.appBoundaryView("Slugbug")
	footer := m.footerView()
	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}

	return tea.NewView(s.Base.Render(header + "\n" + body + "\n\n" + footer))
}

func (m Model) dispatchOperation() (tea.Model, tea.Cmd) {
	switch slugbug.Operation(m.operation) {
	case slugbug.Display:
		m.screen = displayScreen
		return m, displayCmd(m.slugbug)
	case slugbug.Inspect:
		m.screen = resultScreen
		m.result = ""
		m.err = nil
		m.resetResultView()
		return m, serviceListCmd(m.slugbug)
	default:
		m.screen = resultScreen
		m.result = fmt.Sprintf("%s is not implemented yet.", m.operation)
		m.err = nil
		m.resetResultView()
		return m, nil
	}
}

func (m Model) toOperationScreen() (tea.Model, tea.Cmd) {
	m.screen = operationScreen
	m.operation = ""
	m.service = ""
	m.result = ""
	m.err = nil
	m.form = OperationForm(&m.operation)
	return m, m.form.Init()
}

func (m *Model) resetResultView() {
	m.resultView.SetContent(m.result)
	m.resultView.GotoTop()
}

type displayServicesMsg struct {
	services []string
	err      error
}

func displayCmd(sb *slugbug.Slugbug) tea.Cmd {
	return func() tea.Msg {
		services, err := sb.Services()
		return displayServicesMsg{services: services, err: err}
	}
}

func serviceListCmd(sb *slugbug.Slugbug) tea.Cmd {
	return func() tea.Msg {
		services, err := sb.Services()
		return serviceListMsg{services: services, err: err}
	}
}

func inspectCmd(sb *slugbug.Slugbug, service string) tea.Cmd {
	return func() tea.Msg {
		result, err := sb.InspectService(service)
		return operationResultMsg{text: result, err: err}
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

func (m Model) footerView() string {
	statusText := m.styles(m.hasDarkBg).StatusBarText
	name := statusText.Render(m.slugbug.Name())
	if m.operation == "" {
		return m.appBoundaryView(name)
	}

	operation := statusText.Render(m.operation)
	remainingWidth := m.width - lipgloss.Width(name)
	return m.appBoundaryView(name + lipgloss.PlaceHorizontal(
		remainingWidth,
		lipgloss.Right,
		operation,
	))
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
