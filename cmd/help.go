package main

import "github.com/charmbracelet/lipgloss"

func getHelp() string {
	head := "                HELP DEFINITIONS"
	// description := "This utility allows you to manipulate AWS resources easily"
	keySecret := "The AWS Key & Secret information for API access to your AWS environment. Elevated permissions is suggested."
	token := "If you are given rotating session credentials then you will need to enter a session token here. Please note this is not needed if you use the CLI access keys assigned from a static IAM user."
	lambda := "Lambda menu where you can List, Clone & Upgrade Lambda functions. It will upgrade to the latest version of that Runtime. Clone + Upgrade does both actions in 1 shot. Useful for cloning unsupported runtimes in AWS."
	glue := "Glue jobs where you can List, Clone & Upgrade"
	addText := "New text to add to the name of the object that you are cloning. The clone function uses the original name of the selected object and adds whatever text you put in here. It appends this text to the original name. This is a mandatory field to avoid duplicate function entries. For more control on where to add this New text use Replace Text field."
	replaceText := "Text you want to remove and replace with New text. Text entered here will get replaced with the New Text regardless of it's location in the name of the object giving you more control on where to add New Text. If the Replace Text string is not found or if you leave this entry blank then New Text will always default to append to the end of the object."

	finalX := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true).Render(head) + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Key/Secret: ") + keySecret + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Token: ") + token + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Lambda: ") + lambda + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Glue: ") + glue + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("New Text: ") + addText + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("112")).Bold(true).Render("Replace Text: ") + replaceText
	return finalX
}
