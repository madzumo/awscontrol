package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
)

var (
	lipTitleStyle       = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205"))
	itemStyle           = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle     = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle           = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	textPromptColor     = "120" //"100" //nice: 141
	textInputColor      = "140" //"40" //nice: 193
	textResultJob       = "120" //PINK"205"
	spinnerColor        = "226"
	textErrorColorBack  = "1"
	textErrorColorFront = "15"
	textJobOutcomeFront = "216"

	menuTOP = []string{
		"Enter AWS Key",
		"Enter AWS Secret",
		"Enter Region",
		"Enter Session Token",
		"Clone Lambda - Enter Function",
		"Upgrade Lambda - Enter Function",
		"Save Settings",
	}
)

// App States
type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateResultDisplay
	StateSpinner
	StateTextInput
)

// Messsage returend when the background job finishes
type backgroundJobMsg struct {
	result string
}

// // message returned when you have to continue the prompting of data
// type continueLambda struct {
// 	result string
// }

type JobList int

type MenuList struct {
	list                list.Model
	choice              string
	header              string
	state               MenuState
	prevState           MenuState
	prevMenuState       MenuState
	spinner             spinner.Model
	spinnerMsg          string
	backgroundJobResult string
	textInput           textinput.Model
	inputPrompt         string
	textInputError      bool
	jobOutcome          string
	app                 *applicationMain
	lambdaFunction      string
}

func (m MenuList) Init() tea.Cmd {
	return nil
}

func (m MenuList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMainMenu:
		return m.updateMainMenu(msg)
	case StateSpinner:
		return m.updateSpinner(msg)
	case StateTextInput:
		return m.updateTextInput(msg)
	case StateResultDisplay:
		return m.updateResultDisplay(msg)
	default:
		return m, nil
	}
}

func (m *MenuList) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// case tea.MouseMsg:
	// 	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
	// 		err := clipboard.WriteAll(m.headerIP)
	// 		if err != nil {
	// 			fmt.Println("Failed to copy to clipboard:", err)
	// 		}
	// 	}
	// 	return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "Q":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				switch m.choice {
				case menuTOP[0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[0]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Key123"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[1]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Secret123"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[2]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[2]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., us-east-1"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
					// case menuTOP[3]:
					// 	m.prevState = m.state
					// 	m.prevMenuState = m.state
					// 	m.state = StateSpinner
					// 	return m, tea.Batch(m.spinner.Tick, m.backgroundCloneLambda())
				case menuTOP[3]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[3] //"Enter Lambda Function Name to Clone"
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Jibberish Characters"
					m.textInput.Focus()
					m.textInput.CharLimit = 1000
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
				case menuTOP[4]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[4] //"Enter Lambda Function Name to Clone"
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Lambda123"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
					// case menuTOP[4]:
					// 	m.prevState = m.state
					// 	m.prevMenuState = m.state
					// 	m.state = StateSpinner
					// 	return m, tea.Batch(m.spinner.Tick, m.backgroundUpdateLambda())
				case menuTOP[5]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[5]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Lambda123"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[6]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundSaveSettings())
				}
			}
			return m, nil
		}
		// case jobListMsg:

		// 	// m.state = StateResultDisplay
		// 	// return m, nil
		// 	m.prevState = m.state
		// 	m.state = StateSpinner
		// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.textInputError = false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := m.textInput.Value() // User pressed enter, save the input

			switch m.inputPrompt {
			case menuTOP[0]:
				m.app.AwsKey = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved AWS Key: %s", wordwrap.WrapString(inputValue, 90))
			case menuTOP[1]:
				m.app.AwsSecret = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved AWS Secret: %s", wordwrap.WrapString(inputValue, 90))
			case menuTOP[2]:
				m.app.Region = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved Region: %s", wordwrap.WrapString(inputValue, 90))
			case menuTOP[3]:
				m.app.SessionToken = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved Session Token: %s", wordwrap.WrapString(inputValue, 90))
			case menuTOP[4]:
				m.lambdaFunction = inputValue
				m.backgroundJobResult = "Starting Lambda Cloning..."

				// Transition to Spinner State
				m.prevState = m.state
				m.state = StateSpinner

				// Run the spinner and background job
				return m, tea.Batch(m.spinner.Tick, m.backgroundCloneLambda(m.lambdaFunction))
			case menuTOP[5]:
				m.lambdaFunction = inputValue
				m.backgroundJobResult = "Starting Lambda Updating..."

				// Transition to Spinner State
				m.prevState = m.state
				m.state = StateSpinner

				// Run the spinner and background job
				return m, tea.Batch(m.spinner.Tick, m.backgroundUpdateLambda(m.lambdaFunction))
			}

			m.prevState = m.state
			m.state = StateResultDisplay
			return m, nil

		case tea.KeyEsc:
			// m.state = StateSettingsMenu
			m.state = m.prevState
			return m, nil
		}
	}

	return m, cmd
}

func (m *MenuList) updateSpinner(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "q", "esc":
		// 	m.backgroundJobResult = "Job Cancelled"
		// 	m.state = StateResultDisplay
		// 	return m, nil
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		m.backgroundJobResult = m.jobOutcome + "\n\n" + msg.result + "\n"
		m.state = StateResultDisplay
		return m, nil
	// case continueLambda:
	// 	return m, tea.Batch(m.spinner.Tick, m.backgroundCloneLambda(m.lambdaFunction))
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *MenuList) updateResultDisplay(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.textInputError {
				m.state = m.prevState
			} else {
				m.state = m.prevMenuState
			}
			m.updateListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuList) viewResultDisplay() string {
	outro := "Press 'esc' to return."
	outroRender := lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true).Render(outro)
	lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true)
	if m.textInputError {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textErrorColorFront)).Background(lipgloss.Color(textErrorColorBack)).Bold(true).Render(m.backgroundJobResult)
	} else {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textResultJob)).Render(m.backgroundJobResult)
	}
	return fmt.Sprintf("\n\n%s\n\n%s", wordwrap.WrapString(m.backgroundJobResult, 90), wordwrap.WrapString(outroRender, 90))

	// //repeat interval
	// if m.configSettings.Interval > 0 {

	// }
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu, StateSettingsMenu:
		m.header = m.app.getHeader()
		return m.header + "\n" + m.list.View()
	case StateSpinner:
		return m.viewSpinner()
	case StateTextInput:
		return m.viewTextInput()
	case StateResultDisplay:
		return m.viewResultDisplay()
	default:
		return "Unknown state"
	}
}

func (m MenuList) viewSpinner() string {
	// tea.ClearScreen()
	spinnerBase := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.spinnerMsg)

	// return spinnerBase + m.jobOutcome
	return spinnerBase + lipgloss.NewStyle().Foreground(lipgloss.Color(textJobOutcomeFront)).Bold(true).Render(m.jobOutcome)
}

func (m MenuList) viewTextInput() string {
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor)).Bold(true)
	return fmt.Sprintf("\n\n%s\n\n%s", promptStyle.Render(m.inputPrompt), m.textInput.View())

}

func (m *MenuList) updateListItems() {
	switch m.state {
	case StateMainMenu:
		items := []list.Item{}
		for _, value := range menuTOP {
			items = append(items, item(value))
		}
		m.list.SetItems(items)
		// case StateSettingsMenu:
		// 	items := []list.Item{}
		// 	for _, value := range menuSettings {
		// 		items = append(items, item(value[0]))
		// 	}
		// 	m.list.SetItems(items)
	}

	m.list.ResetSelected()
}

func (m *MenuList) backgroundSaveSettings() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(spinnerColor)) //white = 231
		m.spinnerMsg = "Saving Settings"
		// m.spinner.Tick()
		m.app.saveSettings()
		time.Sleep(1 * time.Second)
		return backgroundJobMsg{result: "Settings Saved"}
	}
}

func (m *MenuList) backgroundCloneLambda(lambdaFuncName string) tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(spinnerColor)) //white = 231
		m.spinnerMsg = "Cloning Lambda"
		resultX := "The Lamb is cloned"

		message, err := m.app.cloneLambda(lambdaFuncName, fmt.Sprintf("%s-p313", lambdaFuncName))
		if err != nil {
			resultX = message
		}
		return backgroundJobMsg{result: resultX}
	}
}

func (m *MenuList) backgroundUpdateLambda(lambdaFunctionName string) tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(spinnerColor)) //white = 231
		m.spinnerMsg = "Upgrading Lambda Runtime"
		resultX := "The Lamb is Upgraded"

		message, err := m.app.upgradeLambda(lambdaFunctionName)
		if err != nil {
			resultX = message
		}
		return backgroundJobMsg{result: resultX}
	}
}

func ShowMenu(app *applicationMain) {

	const listWidth = 90
	const listHeight = 14

	// Initialize the list with empty items; items will be set in updateListItems
	l := list.New([]list.Item{}, itemDelegate{}, listWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.Styles.Title = lipTitleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help option

	s := spinner.New()
	s.Spinner = spinner.Pulse

	m := MenuList{
		list:       l,
		header:     app.getHeader(),
		state:      StateMainMenu,
		spinner:    s,
		spinnerMsg: "Action Performing",
		app:        app,
	}

	m.updateListItems()

	m.list.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	)

	//show Menu
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
