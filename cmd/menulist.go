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
	lipTitleStyle       = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("207"))
	itemStyle           = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle     = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle           = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	textPromptColor     = "120"
	textInputColor      = "140"
	textConfirmColor    = "226"
	spinnerColor        = "226"
	textErrorColorBack  = "1"
	textErrorColorFront = "15"
	textJobOutcomeFront = "216"
	menuColorLambda     = "214"
	menuColorMain       = "170"
	menuColorGlue       = "51"

	menuTOP = []string{
		"Enter AWS Key",
		"Enter AWS Secret",
		"Enter Region",
		"Enter Session Token",
		"Set Append Text to Name",
		"Set Replace Text in Name",
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Lambda"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("123")).Render("Glue"),
		// lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Render("Step"),
		"Help",
		"Save Settings",
	}

	menuLAMBDA = []string{
		"List Lambda Functions",
		"Clone Lambda",
		"Upgrade Lambda",
	}

	menuGLUE = []string{
		"coming soon...",
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
	case StateMenuMAIN, StateMenuLAMBDA, StateMenuGLUE:
		str := fmt.Sprintf("%d. %s", index+1, i.displayName)
		if index == lm.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + strings.Join(s, " "))
			}
		}
		fmt.Fprint(w, fn(str))
	case StateLambdaClone, StateLambdaUpgrade:
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

	case StateLambdaList:
		fmt.Fprint(w, i.displayName)
	}
}

// App States via Menu
type MenuState int

const (
	StateMenuMAIN MenuState = iota
	StateSettingsMenu
	StateResultDisplay
	StateSpinner
	StateTextInput
	StateLambdaClone
	StateLambdaUpgrade
	StateLambdaList
	StateMenuLAMBDA
	StateMenuGLUE
)

type OutroDisplayState int

const (
	OutroEsc OutroDisplayState = iota
	OutroEnterClone
	OutroEnterUpdate
)

type backgroundJobMsg struct {
	result  string
	isError bool
}

type JobList int

type MenuList struct {
	list      list.Model
	choice    string
	header    string
	state     MenuState
	prevState MenuState
	// prevMenuState       MenuState
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
	case StateMenuMAIN:
		return m.updateMenuMain(msg)
	case StateMenuLAMBDA:
		return m.updateMenuLambda(msg)
	case StateMenuGLUE:
		return m.updateMenuGlue(msg)
	case StateLambdaList:
		return m.updateLambdaList(msg)
	case StateLambdaClone:
		return m.updateLambdaClone(msg)
	case StateLambdaUpgrade:
		return m.updateLambdaUpgrade(msg)
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

func (m *MenuList) updateLambdaList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
			m.list.Title = "Available Lambda Functions (filtered)"
		}
		switch msg.String() {
		case "esc":
			if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
				m.list.ResetFilter()
				m.list.Title = "Available Lambda Functions"
				m.list.Styles.Title = lipTitleStyle
				return m, nil
			} else {
				m.prevState = m.state
				m.state = StateMenuLAMBDA
				m.fillListItems()
				return m, nil
			}

		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateLambdaClone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
			m.list.Title = "Clone Lambda Functions (filtered)"
		}
		switch msg.String() {
		case "esc":
			if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
				m.list.ResetFilter()
				m.list.Title = "Clone Lambda Functions"
				m.list.Styles.Title = lipTitleStyle
				return m, nil
			} else {
				m.prevState = m.state
				m.state = StateMenuLAMBDA
				m.fillListItems()
				return m, nil
			}

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
			if m.list.FilterState() != list.Filtering {
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

func (m *MenuList) updateLambdaUpgrade(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
			m.list.Title = "Upgrade Lambda Functions (filtered)"
		}
		switch msg.String() {
		case "esc":
			if m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied {
				m.list.ResetFilter()
				m.list.Title = "Upgrade Lambda Functions"
				m.list.Styles.Title = lipTitleStyle
				return m, nil
			} else {
				m.prevState = m.state
				m.state = StateMenuLAMBDA
				m.fillListItems()
				return m, nil
			}

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
			if m.list.FilterState() != list.Filtering {
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

func (m *MenuList) updateMenuMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(*itemX)
			if ok {
				m.choice = string(i.name)
				switch m.choice {
				case menuTOP[0]:
					// m.prevMenuState = m.state
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
					// m.prevMenuState = m.state
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
					// m.prevMenuState = m.state
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
				case menuTOP[3]:
					// m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[3]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Jibberish Characters"
					m.textInput.Focus()
					m.textInput.CharLimit = 1000
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[4]:
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[4]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., -prod"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[5]:
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[5]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., -new-prod"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[6]:
					m.prevState = m.state
					m.list.Title = "Main Menu->Lambda"
					m.state = StateMenuLAMBDA
					m.fillListItems()
					return m, nil
				case menuTOP[7]:
					m.prevState = m.state
					m.state = StateMenuGLUE
					m.fillListItems()
					return m, nil
				case menuTOP[8]:
					m.state = StateResultDisplay
					m.stateOutroDisplay = OutroEsc
					m.backgroundJobResult = getHelp()
					return m, nil
				case menuTOP[9]:
					m.prevState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundSaveSettings())
				}
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateMenuLambda(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = StateMenuMAIN
			m.fillListItems()
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(*itemX)
			if ok {
				m.choice = string(i.name)
				switch m.choice {
				case menuLAMBDA[0]:
					m.prevState = m.state
					m.state = StateLambdaList
					m.fillListItems()
					return m, nil
				case menuLAMBDA[1]:
					if m.app.FileNameExtension == "" {
						m.state = StateResultDisplay
						m.stateOutroDisplay = OutroEsc
						m.backgroundJobResult = "Append Text required to clone"
						m.textInputError = true
						return m, nil
					} else {
						m.prevState = m.state
						m.state = StateLambdaClone
						m.fillListItems()
						return m, nil
					}
				case menuLAMBDA[2]:
					m.prevState = m.state
					m.state = StateLambdaUpgrade
					m.fillListItems()
					return m, nil
				}
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateMenuGlue(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "enter":
			m.state = StateMenuMAIN
			m.fillListItems()
			return m, nil
			// case "enter":
			// 	i, ok := m.list.SelectedItem().(*itemX)
			// 	if ok {
			// 		m.choice = string(i.name)
			// 		switch m.choice {
			// 		case menuLAMBDA[0]:
			// 			m.prevState = m.state
			// 			m.state = StateLambdaList
			// 			m.fillListItems()
			// 			return m, nil
			// 		case menuLAMBDA[1]:
			// 			m.prevState = m.state
			// 			m.state = StateLambdaClone
			// 			m.fillListItems()
			// 			return m, nil
			// 		case menuTOP[2]:
			// 			m.prevState = m.state
			// 			m.state = StateLambdaUpgrade
			// 			m.fillListItems()
			// 			return m, nil
			// 		}
			// 	}
			// 	return m, nil
		}
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
				m.backgroundJobResult = fmt.Sprintf("Saved Append Text to File Name: %s", inputValue)
			case menuTOP[5]:
				m.app.ReplaceExtension = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved Replace Text to File Name: %s", inputValue)
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
		if m.prevState != StateLambdaClone && m.prevState != StateLambdaUpgrade && m.prevState != StateLambdaList {
			m.prevState = m.state
		}
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
			//this requires special conditionals becuase ResultDisplay is used to show
			//results but also for list selection
			if m.prevState == StateLambdaClone || m.prevState == StateLambdaUpgrade || m.prevState == StateLambdaList {
				m.state = StateMenuLAMBDA
			} else {
				m.state = StateMenuMAIN
			}
			m.fillListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.prevState == StateLambdaClone {
				m.state = StateSpinner
				return m, tea.Batch(m.spinner.Tick, m.backgroundCloneLambda())
			} else if m.prevState == StateLambdaUpgrade {
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
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textErrorColorFront)).Background(lipgloss.Color(textErrorColorBack)).Bold(true).Render(m.backgroundJobResult)
	} else {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textConfirmColor)).Render(m.backgroundJobResult)
	}

	return fmt.Sprintf("\n\n%s\n\n%s", wordwrap.WrapString(m.backgroundJobResult, 90), outroRender)
}

func (m MenuList) View() string {
	switch m.state {
	case StateMenuMAIN, StateMenuLAMBDA, StateMenuGLUE:
		m.header = m.app.getHeader()
		return m.header + "\n" + m.list.View()
	case StateLambdaClone, StateLambdaUpgrade, StateLambdaList:
		return m.list.View()
	case StateSpinner:
		return m.viewSpinner()
	case StateTextInput:
		return m.viewTextInput()
	case StateResultDisplay:
		return m.viewResultDisplay()
	default:
		return "Unknown State"
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
	case StateMenuMAIN:
		items := []list.Item{}
		for _, value := range menuTOP {
			items = append(items, &itemX{value, false, value})
		}
		m.list.SetItems(items)

	case StateMenuLAMBDA:
		items := []list.Item{}
		for _, value := range menuLAMBDA {
			items = append(items, &itemX{value, false, value})
		}
		m.list.SetItems(items)

	case StateMenuGLUE:
		items := []list.Item{}
		for _, value := range menuGLUE {
			items = append(items, &itemX{value, false, value})
		}
		m.list.SetItems(items)

	case StateLambdaClone, StateLambdaUpgrade, StateLambdaList:
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
			if m.app.ReplaceExtension != "" {
				if strings.Contains(v, m.app.ReplaceExtension) {
					newNameX = strings.Replace(v, m.app.ReplaceExtension, m.app.FileNameExtension, -1)
				} else {
					newNameX = fmt.Sprintf("%s%s", v, m.app.FileNameExtension)
				}
			} else {
				newNameX = fmt.Sprintf("%s%s", v, m.app.FileNameExtension)
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
			err := m.app.upgradeLambda(v)
			if err != nil {
				resultX = err.Error()
				continue
			}
		}
		return backgroundJobMsg{result: resultX}
	}
}

func SetupListMenu(currentState MenuState) list.Model {
	listWidth := 90
	listHeight := 12

	// Initialize the list with empty items; items will be set in FillListItems
	lm := list.New([]list.Item{}, itemDelegateX{currentState: currentState}, listWidth, listHeight)
	lm.SetShowStatusBar(false)
	switch currentState {
	case StateMenuMAIN:
		lm.SetFilteringEnabled(false)
		lm.SetShowTitle(false)
		selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuColorMain))
	case StateMenuLAMBDA:
		lm.SetFilteringEnabled(false)
		lm.SetShowTitle(false)
		selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuColorLambda))
	case StateMenuGLUE:
		lm.SetFilteringEnabled(false)
		lm.SetShowTitle(false)
		selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuColorGlue))
	case StateLambdaClone:
		lm.SetHeight(27)
		lm.SetFilteringEnabled(true)
		lm.SetShowTitle(true)
		lm.Title = "Clone Lambda Functions"
	case StateLambdaUpgrade:
		lm.SetHeight(27)
		lm.SetFilteringEnabled(true)
		lm.SetShowTitle(true)
		lm.Title = "Upgrade Lambda Functions"
	case StateLambdaList:
		lm.SetHeight(27)
		lm.SetFilteringEnabled(true)
		lm.SetShowTitle(true)
		lm.Title = "Available Lambda Functions"
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
		state:      StateMenuMAIN,
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
