package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func (app *applicationMain) createLambdaClient() (*lambda.Client, error) {
	ctx := context.Background()
	customCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(app.AwsKey, app.AwsSecret, app.SessionToken),
	)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(customCreds), config.WithRegion(app.Region))
	if err != nil {
		return nil, err
	}

	client := lambda.NewFromConfig(cfg)
	return client, nil
}

func (app *applicationMain) cloneLambda(functionName string, functionNameNew string) (string, error) {
	ctx := context.Background()
	//create lambda client
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		return fmt.Sprintf("Failed to create Lambda connection:\n%v", err), err
	}

	//get the lambda function
	result, err := clientLamb.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Sprintf("Failed to get function details:\n%v", err), err
	}

	//download the lambda Zip file
	if result.Code == nil || result.Code.Location == nil {
		return "No code location found for the function", fmt.Errorf("missing code location")
	}
	resp, err := http.Get(*result.Code.Location)
	if err != nil {
		return fmt.Sprintf("Failed to download code function:\n%v", err), err
	}
	defer resp.Body.Close()

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Failed to read code zip file content:\n%v", err), err
	}

	//get layer information
	var layerArns []string
	if result.Configuration.Layers != nil {
		for _, layer := range result.Configuration.Layers {
			if layer.Arn != nil {
				layerArns = append(layerArns, *layer.Arn)
			}
		}
	}

	//get environment
	var env *types.Environment
	if result.Configuration.Environment != nil {
		env = &types.Environment{
			Variables: result.Configuration.Environment.Variables,
		}
	}

	//create the new lambda
	_, err = clientLamb.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionNameNew),
		Runtime:      result.Configuration.Runtime,
		Role:         result.Configuration.Role,
		Handler:      result.Configuration.Handler,
		Code: &types.FunctionCode{
			ZipFile: zipBytes,
		},
		Timeout:       result.Configuration.Timeout,
		MemorySize:    result.Configuration.MemorySize,
		Environment:   env,
		Layers:        layerArns,
		TracingConfig: (*types.TracingConfig)(result.Configuration.TracingConfig),
		Architectures: result.Configuration.Architectures,
		PackageType:   result.Configuration.PackageType,
		Description:   result.Configuration.Description,
		Publish:       *aws.Bool(true),
	})
	if err != nil {
		return fmt.Sprintf("Failed to create a new lambda function:\n%v", err), err
	}

	return "", nil
}

func (app *applicationMain) upgradeLambda(lambdaFunctionName string) (string, error) {
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		return fmt.Sprintf("Failed to create Lambda connection:\n%v", err), err
	}

	newRuntime := types.RuntimePython313

	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(lambdaFunctionName),
		Runtime:      newRuntime,
	}

	_, err = clientLamb.UpdateFunctionConfiguration(context.TODO(), input)
	if err != nil {
		return fmt.Sprintf("Failed to udpate Lambda function\n%v", err), err
	}

	return "", nil
}

func (app *applicationMain) listAllLambdaFunctions() (LambdaItems [][]string, err error) {
	ctx := context.Background()
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		LambdaItems = append(LambdaItems, []string{fmt.Sprintf("Failed to create Lambda connection:\n%v", err), ""})
		return LambdaItems, err
	}

	resp, err := clientLamb.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		LambdaItems = append(LambdaItems, []string{fmt.Sprintf("Failed to list Lambda functions:\n%v", err), ""})
		return LambdaItems, err
	}
	// fmt.Println("Available Lambda Functions:")
	for _, fn := range resp.Functions {
		LambdaItems = append(LambdaItems, []string{aws.ToString(fn.FunctionName), string(fn.Runtime)})
	}

	return LambdaItems, nil
}
