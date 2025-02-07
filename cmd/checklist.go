package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type itemLamb struct {
	name     string
	selected bool
}

func (i itemLamb) FilterValue() string { return i.name }

type itemDelegateLamb struct{}

func (d itemDelegateLamb) Height() int                             { return 1 }
func (d itemDelegateLamb) Spacing() int                            { return 0 }
func (d itemDelegateLamb) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegateLamb) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(itemLamb)
	if !ok {
		return
	}

	checkbox := "[ ]"
	if i.selected {
		checkbox = "[x]"
	}

	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}
	str := fmt.Sprintf("%s%s %s", cursor, checkbox, i.name)

	// fn := itemStyle.Render
	// if index == m.Index() {
	// 	fn = func(s ...string) string {
	// 		return selectedItemStyle.Render("> " + strings.Join(s, " "))
	// 	}
	// }

	fmt.Fprint(w, str)
}
