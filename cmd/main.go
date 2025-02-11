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
					AWS Control 1.4
`

	settingsFileName = "settings.json"
	headerColor      = "170"
	subHeaderColor   = "120"
	// fileNameFiller   = "p313"
)

type applicationMain struct {
	AwsKey            string `json:"awskey"`
	AwsSecret         string `json:"awssecret"`
	Region            string `json:"region"`
	SessionToken      string `json:"session"`
	FileNameExtension string `json:"filenameextension"`
	ReplaceExtension  string `json:"replaceextension"`
}

func main() {
	app := &applicationMain{AwsKey: "-", AwsSecret: "-", Region: "-"}
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

func (app *applicationMain) getHeader() string {
	fullHeader := lipgloss.NewStyle().Foreground(lipgloss.Color(headerColor)).Render(headerMenu) + "\n" +
		fmt.Sprintf("         Key: %s...  Region: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.truncateX(app.AwsKey, 10)),
			lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.Region)) +
		fmt.Sprintf("      Secret: %s...   Token: %s...\n", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.truncateX(app.AwsSecret, 10)),
			lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.truncateX(app.SessionToken, 10))) +
		fmt.Sprintf("    New Text: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.FileNameExtension)) +
		fmt.Sprintf("Replace Text: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.ReplaceExtension))

	return fullHeader
}

func (app *applicationMain) saveSettings() {
	data, err := json.MarshalIndent(app, "", " ")
	if err != nil {
		fmt.Printf("Error saving settings\n%s", err)
	}

	err = os.WriteFile(settingsFileName, data, 0644)
	if err != nil {
		fmt.Printf("Error saving settings\n%s", err)
	}
}

func (app *applicationMain) truncateX(s string, length int) string {
	runes := []rune(s) // Convert string to runes to handle Unicode properly
	if len(runes) > length {
		return string(runes[:length])
	}
	return s
}
