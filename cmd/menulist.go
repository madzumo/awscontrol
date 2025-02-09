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
	lipTitleStyle       = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("45"))
	lipTitleStyleFilter = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("207"))
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
		"Set Name Extension",
		"Clone Lambda",
		"Upgrade Lambda",
		"Save Settings",
	}
)

// MENU STRUCTURE
type itemX struct {
	name        string
	selected    bool
	displayName string
}

func (i *itemX) FilterValue() string { return i.displayName }

type itemDelegateX struct {
	currentState MenuState
}

func (d itemDelegateX) Height() int                             { return 1 }
func (d itemDelegateX) Spacing() int                            { return 0 }
func (d itemDelegateX) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegateX) Render(w io.Writer, lm list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*itemX)
	if !ok {
		return
	}

	fn := itemStyle.Render

	switch d.currentState {
	case StateMainMenu:
		str := fmt.Sprintf("%d. %s", index+1, i.displayName)
		if index == lm.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + strings.Join(s, " "))
			}
		}
		fmt.Fprint(w, fn(str))
	case StateLambdaClone, StateLambdaUpdate:
		checkbox := "[ ]"
		if i.selected {
			checkbox = "[x]"
		}
		cursor := "  "
		if index == lm.Index() {
			cursor = "> "
		}
		str := fmt.Sprintf("%s%s %s", cursor, checkbox, i.displayName)
		fmt.Fprint(w, str)
	}
}

// App States via Menu
type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateResultDisplay
	StateSpinner
	StateTextInput
	StateLambdaClone
	StateLambdaUpdate
)

type OutroDisplayState int

const (
	OutroEsc OutroDisplayState = iota
	OutroEnterClone
	OutroEnterUpdate
)

type backgroundJobMsg struct {
	result string
}

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
	stateOutroDisplay   OutroDisplayState
	lambdaSelectedList  []string
}

func (m MenuList) Init() tea.Cmd {
	return nil
}

func (m MenuList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMainMenu:
		return m.updateMainMenu(msg)
	case StateLambdaClone:
		return m.updateLambdaClone(msg)
	case StateLambdaUpdate:
		return m.updateLambdaUpdate(msg)
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

func (m *MenuList) updateLambdaClone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			switch msg.String() {
			case "esc":
				m.list.ResetFilter()
				m.list.Title = "Available Lambda Functions - Clone"
				m.list.Styles.Title = lipTitleStyle
				return m, nil
			}
			m.list.Title = "Available Lambda Functions - Clone(filtered)"
			m.list.Styles.Title = lipTitleStyleFilter
		} else {
			switch msg.String() {
			case "esc", "Q", "q":
				m.prevMenuState = m.state
				m.prevState = m.state
				m.state = StateMainMenu
				m.fillListItems()
				return m, nil
			case " ":
				i, ok := m.list.SelectedItem().(*itemX)
				if ok {
					for idx, val := range m.list.Items() {
						item := val.(*itemX)
						if item.name == i.name {
							item.selected = !item.selected
							m.list.SetItem(idx, item)
						}
					}
				}
			case "enter":
				selectedItems := []string{}
				for _, it := range m.list.Items() {
					i := it.(*itemX)
					if i.selected {
						selectedItems = append(selectedItems, i.name)
					}
				}
				if len(selectedItems) > 0 {
					m.lambdaSelectedList = selectedItems
					m.backgroundJobResult = strings.Join(selectedItems, "\n")
					m.prevState = m.state
					m.stateOutroDisplay = OutroEnterClone
					m.state = StateResultDisplay
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateLambdaUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			switch msg.String() {
			case "esc":
				m.list.ResetFilter()
				m.list.Title = "Available Lambda Functions - Upgrade"
				m.list.Styles.Title = lipTitleStyle
				return m, nil
			}
			m.list.Title = "Available Lambda Functions - Upgrade(filtered)"
			m.list.Styles.Title = lipTitleStyleFilter
		} else {
			switch msg.String() {
			case "esc", "Q", "q":
				m.prevMenuState = m.state
				m.prevState = m.state
				m.state = StateMainMenu
				m.fillListItems()
				return m, nil
			case " ":
				i, ok := m.list.SelectedItem().(*itemX)
				if ok {
					for idx, val := range m.list.Items() {
						item := val.(*itemX)
						if item.name == i.name {
							item.selected = !item.selected
							m.list.SetItem(idx, item)
						}
					}
				}
			case "enter":
				selectedItems := []string{}
				for _, it := range m.list.Items() {
					i := it.(*itemX)
					if i.selected {
						selectedItems = append(selectedItems, i.name)
					}
				}
				if len(selectedItems) > 0 {
					m.lambdaSelectedList = selectedItems
					m.backgroundJobResult = strings.Join(selectedItems, "\n")
					m.prevState = m.state
					m.stateOutroDisplay = OutroEnterClone
					m.state = StateResultDisplay
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "Q":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(*itemX)
			if ok {
				m.choice = string(i.name)
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
					m.inputPrompt = menuTOP[4]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., -dev"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[5]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateLambdaClone
					m.fillListItems()
					return m, nil
				case menuTOP[6]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateLambdaUpdate
					m.fillListItems()
					return m, nil
				case menuTOP[7]:
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
				m.backgroundJobResult = fmt.Sprintf("Saved AWS Key: %s", inputValue)
			case menuTOP[1]:
				m.app.AwsSecret = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved AWS Secret: %s", inputValue)
			case menuTOP[2]:
				m.app.Region = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved Region: %s", inputValue)
			case menuTOP[3]:
				m.app.SessionToken = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved Session Token: %s", inputValue)
			case menuTOP[4]:
				m.app.FileNameExtension = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved File Name Extension: %s", inputValue)
			}

			m.prevState = m.state
			m.stateOutroDisplay = OutroEsc
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
		m.prevState = m.state
		m.stateOutroDisplay = OutroEsc
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
			if m.prevState == StateLambdaClone || m.prevState == StateLambdaUpdate {
				m.state = m.prevState
			} else {
				m.state = m.prevMenuState
			}
			m.fillListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.prevState == StateLambdaClone {
				// m.prevState = m.state
				m.state = StateSpinner
				return m, tea.Batch(m.spinner.Tick, m.backgroundCloneLambda())
			} else if m.prevState == StateLambdaUpdate {
				// m.prevState = m.state
				m.state = StateSpinner
				return m, tea.Batch(m.spinner.Tick, m.backgroundUpdateLambda())
			}
		}
	}
	return m, nil
}

func (m MenuList) viewResultDisplay() string {

	var outro string
	switch m.stateOutroDisplay {
	case OutroEsc:
		outro = "Press 'esc' to return."
	case OutroEnterClone:
		outro = "Press 'enter' to Clone these Lambda functions"
	case OutroEnterUpdate:
		outro = "Press 'enter' to Upgrade these Lambda functions"
	}

	outroRender := lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true).Render(outro)
	if m.textInputError {
		m.backgroundJobResult = wordwrap.WrapString(lipgloss.NewStyle().Foreground(lipgloss.Color(textErrorColorFront)).Background(lipgloss.Color(textErrorColorBack)).Bold(true).Render(m.backgroundJobResult), 90)
	} else {
		m.backgroundJobResult = wordwrap.WrapString(lipgloss.NewStyle().Foreground(lipgloss.Color(textResultJob)).Render(m.backgroundJobResult), 90)
	}

	return fmt.Sprintf("\n\n%s\n\n%s", m.backgroundJobResult, outroRender)
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu:
		m.header = m.app.getHeader()
		return m.header + "\n" + m.list.View()
	case StateLambdaClone, StateLambdaUpdate:
		return m.list.View()
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

func (m *MenuList) fillListItems() {
	m.list = SetupListMenu(m.state)
	switch m.state {
	case StateMainMenu:
		items := []list.Item{}
		for _, value := range menuTOP {
			items = append(items, &itemX{value, false, value})
		}
		m.list.SetItems(items)
	case StateLambdaClone, StateLambdaUpdate:
		lambdas, err := m.app.listAllLambdaFunctions()
		if err != nil {
			m.backgroundJobResult = err.Error()
			m.stateOutroDisplay = OutroEsc
			m.state = StateResultDisplay
			break
		}
		items := []list.Item{}
		for _, value := range lambdas {
			items = append(items, &itemX{value[0], false, fmt.Sprintf("%s   %s", value[0], lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(value[1]))})
		}

		m.list.SetItems(items)
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

func (m *MenuList) backgroundCloneLambda() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(spinnerColor)) //white = 231
		m.spinnerMsg = "Cloning Lambda"
		resultX := "The Lamb is Cloned"

		var newNameX string
		for _, v := range m.lambdaSelectedList {
			if strings.Contains(v, m.app.FileNameExtension) {
				newNameX = fmt.Sprintf("%s-py313%s", strings.Replace(v, m.app.FileNameExtension, "", 1), m.app.FileNameExtension)
			} else {
				newNameX = fmt.Sprintf("%s-py313%s", v, m.app.FileNameExtension)
			}
			err := m.app.cloneLambda(v, newNameX)
			if err != nil {
				resultX = err.Error()
				continue
			}
		}
		return backgroundJobMsg{result: resultX}
	}
}

func (m *MenuList) backgroundUpdateLambda() tea.Cmd {
	return func() tea.Msg {
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(spinnerColor)) //white = 231
		m.spinnerMsg = "Upgrading Lambda Runtime"
		resultX := "The Lamb is Upgraded"

		for _, v := range m.lambdaSelectedList {
			message, err := m.app.upgradeLambda(v)
			if err != nil {
				resultX = message
				continue
			}
		}
		return backgroundJobMsg{result: resultX}
	}
}

func SetupListMenu(currentState MenuState) list.Model {
	listWidth := 90
	listHeight := 12

	// Initialize the list with empty items; items will be set in updateListItems
	lm := list.New([]list.Item{}, itemDelegateX{currentState: currentState}, listWidth, listHeight)
	lm.SetShowStatusBar(false)
	if currentState == StateMainMenu {
		lm.SetFilteringEnabled(false)
		lm.SetShowTitle(false)
	} else if currentState == StateLambdaClone {
		lm.SetHeight(27)
		lm.SetFilteringEnabled(true)
		lm.SetShowTitle(true)
		lm.Title = "Available Lambda Functions - Clone"
	} else if currentState == StateLambdaUpdate {
		lm.SetHeight(27)
		lm.SetFilteringEnabled(true)
		lm.SetShowTitle(true)
		lm.Title = "Available Lambda Functions - Upgrade"
	}
	lm.Styles.Title = lipTitleStyle
	lm.Styles.PaginationStyle = paginationStyle
	lm.Styles.HelpStyle = helpStyle
	lm.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help option
	lm.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	)
	return lm
}

func ShowMenu(app *applicationMain) {

	s := spinner.New()
	s.Spinner = spinner.Pulse

	m := MenuList{
		header:     app.getHeader(),
		state:      StateMainMenu,
		spinner:    s,
		spinnerMsg: "Action Performing",
		app:        app,
	}

	m.fillListItems()

	//start Bubbles loop
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
