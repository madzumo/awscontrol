package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerMenu = `
      ____           
     /___/\_          
    _\   \/_/\__        
  __\       \/_/\       
  \   __    __ \ \       
 __\  \_\   \_\ \ \   __ 
/_/\\   __   __  \ \_/_/\
\_\/_\__\/\__\/\__\/_\_\/
   \_\/_/\       /_\_\/  
      \_\/       \_\/    
               AWS Control
`

	settingsFileName = "settings.json"
	headerColor      = "170"
	fileNameFiller   = "p313"
)

type applicationMain struct {
	AwsKey    string `json:"awskey"`
	AwsSecret string `json:"awssecret"`
}

func main() {
	app := &applicationMain{}
	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		fmt.Printf("Error getting settings\n%s", err)
	}
	err = json.Unmarshal(data, &app)
	if err != nil {
		fmt.Printf("Error getting settings\n%s", err)
	}

	ShowMenu(app)
}

func (a *applicationMain) getHeader() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(headerColor)).Render(headerMenu)
}

func (a *applicationMain) saveSettings() {
	data, err := json.MarshalIndent(a, "", " ")
	if err != nil {
		fmt.Printf("Error saving settings\n%s", err)
	}

	err = os.WriteFile(settingsFileName, data, 0644)
	if err != nil {
		fmt.Printf("Error saving settings\n%s", err)
	}
}

func (a *applicationMain) cloneLambda() {

}

func (a *applicationMain) upgradeLambda() {

}
