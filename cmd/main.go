package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
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
	subHeaderColor   = "120"
	fileNameFiller   = "p313"
)

type applicationMain struct {
	AwsKey    string `json:"awskey"`
	AwsSecret string `json:"awssecret"`
	Region    string `json:"region"`
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

func (app *applicationMain) getHeader() string {
	fullHeader := lipgloss.NewStyle().Foreground(lipgloss.Color(headerColor)).Render(headerMenu) + "\n" +
		fmt.Sprintf("Key: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.AwsKey)) +
		fmt.Sprintf("Secret: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.AwsSecret)) +
		fmt.Sprintf("Region: %s", lipgloss.NewStyle().Foreground(lipgloss.Color(subHeaderColor)).Render(app.Region))

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

func (app *applicationMain) cloneLambda() {
	ctx := context.Background()
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		fmt.Printf("Failed to create Lambda connection:\n%v", err)
	}

	resp, err := clientLamb.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		fmt.Printf("Failed to list Lambda functions:\n%v", err)
	}
	fmt.Println("Available Lambda Functions:")
	for _, fn := range resp.Functions {
		fmt.Println(*fn.FunctionName)
	}
}

func (app *applicationMain) upgradeLambda() {

}

func (app *applicationMain) createLambdaClient() (*lambda.Client, error) {
	ctx := context.Background()
	customCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(app.AwsKey, app.AwsSecret, ""),
	)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(customCreds), config.WithRegion(app.Region))
	if err != nil {
		return nil, err
	}

	client := lambda.NewFromConfig(cfg)
	return client, nil
}
