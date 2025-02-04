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
	fileNameFiller   = "p313"
)

type applicationMain struct {
	AwsKey    string `json:"awskey"`
	AwsSecret string `json:"awssecret"`
	Region    string `json: "region"`
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
	ctx := context.Background()
	clientLamb, err := a.createLambdaClient()
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

func (a *applicationMain) upgradeLambda() {

}

func (a *applicationMain) createLambdaClient() (*lambda.Client, error) {
	ctx := context.Background()
	customCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(a.AwsKey, a.AwsSecret, ""),
	)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(customCreds), config.WithRegion(a.Region))
	if err != nil {
		return nil, err
	}

	client := lambda.NewFromConfig(cfg)
	return client, nil
}
