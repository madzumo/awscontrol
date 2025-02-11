package main

import "github.com/charmbracelet/lipgloss"

func getHelp() string {
	head := "                HELP DEFINITIONS"
	// description := "This utility allows you to manipulate AWS resources easily"
	keySecret := "The AWS Key & Secret information for API access to your AWS environment. Elevated permissions is suggested."
	token := "If you are given rotating session credentials then you will need to enter a session token here. Please note this is not needed if you use the CLI access keys assigned from a static IAM user."
	lambda := "Lambda menu where you can List, Clone & Upgrade Lambda functions. It will upgrade to the latest version of that Runtime. Clone + Upgrade does both actions in 1 shot. Useful for cloning unsupported runtimes in AWS."
	glue := "Glue jobs where you can List, Clone & Upgrade"
	appendText := "Add custom text to your cloned AWS objects. The clone function uses the original name of the selected object and appends whatever text you put in here. This is a mandatory field to avoid duplicate function entries."
	replaceText := "If you want more control on where to place the Append text then use this setting. Whatever text you enter here will get replaced with the Append Text regardless of it's location in the name of the object. If the target Replace Text is not found then the default behavior of appending to the end of the object name will be applied. The same default behavior will apply if you leave this field blank."

	finalX := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true).Render(head) + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Key/Secret: ") + keySecret + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Token: ") + token + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Lambda: ") + lambda + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Glue: ") + glue + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Append Text: ") + appendText + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Replace Text: ") + replaceText
	return finalX
}
