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

func (app *applicationMain) cloneLambda(functionName string, functionNameNew string, upgrade2 bool) error {
	ctx := context.Background()
	//create lambda client
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		return fmt.Errorf("failed to create Lambda connection:\n%v", err)
	}

	//get the lambda function
	result, err := clientLamb.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to get function details:\n%v", err)
	}

	//download the lambda Zip file
	if result.Code == nil || result.Code.Location == nil {
		return fmt.Errorf("no code location found for the function")
	}
	resp, err := http.Get(*result.Code.Location)
	if err != nil {
		return fmt.Errorf("failed to download code function:\n%v", err)
	}
	defer resp.Body.Close()

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read code zip file content:\n%v", err)
	}

	//layers
	var layerArns []string
	if result.Configuration.Layers != nil {
		for _, layer := range result.Configuration.Layers {
			if layer.Arn != nil {
				layerArns = append(layerArns, *layer.Arn)
			}
		}
	}

	//environment variables
	var env *types.Environment
	if result.Configuration.Environment != nil {
		env = &types.Environment{
			Variables: result.Configuration.Environment.Variables,
		}
	}

	//vpc config
	var vpcConfig *types.VpcConfig
	if result.Configuration.VpcConfig != nil {
		vpcConfig = &types.VpcConfig{
			SecurityGroupIds: result.Configuration.VpcConfig.SecurityGroupIds,
			SubnetIds:        result.Configuration.VpcConfig.SubnetIds,
		}
	}

	//runtime selection
	runtimeToUse := result.Configuration.Runtime
	if upgrade2 {
		runtimeToUse = types.RuntimePython313
	}
	//create the new lambda
	newLamb, err := clientLamb.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionNameNew),
		Runtime:      runtimeToUse,
		Role:         result.Configuration.Role,
		Handler:      result.Configuration.Handler,
		Code: &types.FunctionCode{
			ZipFile: zipBytes,
		},
		Timeout:           result.Configuration.Timeout,
		MemorySize:        result.Configuration.MemorySize,
		Environment:       env,
		Layers:            layerArns,
		TracingConfig:     (*types.TracingConfig)(result.Configuration.TracingConfig),
		Architectures:     result.Configuration.Architectures,
		PackageType:       result.Configuration.PackageType,
		Description:       result.Configuration.Description,
		Publish:           *aws.Bool(true),
		VpcConfig:         vpcConfig,
		DeadLetterConfig:  result.Configuration.DeadLetterConfig,
		FileSystemConfigs: result.Configuration.FileSystemConfigs,
		EphemeralStorage:  result.Configuration.EphemeralStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to create a new lambda function:\n%v", err)
	}

	//copy tags
	tagResp, err := clientLamb.ListTags(ctx, &lambda.ListTagsInput{
		Resource: result.Configuration.FunctionArn,
	})
	if err == nil && len(tagResp.Tags) > 0 {
		_, err = clientLamb.TagResource(ctx, &lambda.TagResourceInput{
			Resource: newLamb.FunctionArn,
			Tags:     tagResp.Tags,
		})
		if err != nil {
			return fmt.Errorf("failed to add tags to new Lambda function: %v", err)
		}
	}

	// Copy Concurrency (if set).
	concurrencyResp, err := clientLamb.GetFunctionConcurrency(ctx, &lambda.GetFunctionConcurrencyInput{
		FunctionName: aws.String(functionName),
	})
	if err == nil && concurrencyResp.ReservedConcurrentExecutions != nil {
		_, err = clientLamb.PutFunctionConcurrency(ctx, &lambda.PutFunctionConcurrencyInput{
			FunctionName:                 aws.String(functionNameNew),
			ReservedConcurrentExecutions: concurrencyResp.ReservedConcurrentExecutions,
		})
		if err != nil {
			return fmt.Errorf("failed to set concurrency on new Lambda function: %v", err)
		}
	}

	// Copy Event Source Mappings.
	eventSrcResp, err := clientLamb.ListEventSourceMappings(ctx, &lambda.ListEventSourceMappingsInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to list event source mappings: %v", err)
	}

	for _, src := range eventSrcResp.EventSourceMappings {
		// Determine if the mapping is enabled by comparing its State to "Enabled"
		enabled := false
		if src.State != nil && *src.State == "Enabled" {
			enabled = true
		}
		_, err = clientLamb.CreateEventSourceMapping(ctx, &lambda.CreateEventSourceMappingInput{
			EventSourceArn:                 src.EventSourceArn,
			FunctionName:                   aws.String(functionNameNew),
			BatchSize:                      src.BatchSize,
			Enabled:                        aws.Bool(enabled),
			MaximumBatchingWindowInSeconds: src.MaximumBatchingWindowInSeconds,
			StartingPosition:               src.StartingPosition, // if applicable for your event source
			// Add other necessary fields here.
		})
		if err != nil {
			return fmt.Errorf("failed to create event source mapping on new Lambda function: %v", err)
		}
	}

	//resource policies
	policyResp, err := clientLamb.GetPolicy(ctx, &lambda.GetPolicyInput{
		FunctionName: aws.String(functionName),
	})
	if err == nil && policyResp.Policy != nil {
		_, err = clientLamb.AddPermission(ctx, &lambda.AddPermissionInput{
			FunctionName: newLamb.FunctionArn,
			StatementId:  aws.String("ClonePermission"),
			Action:       aws.String("lambda:InvokeFunction"),
			Principal:    aws.String("*"),
			SourceArn:    newLamb.FunctionArn, // Adjust as needed.
		})
		if err != nil {
			return fmt.Errorf("failed to add permission on new Lambda function: %v", err)
		}
	}

	//aliases
	aliasResp, err := clientLamb.ListAliases(ctx, &lambda.ListAliasesInput{
		FunctionName: aws.String(functionName),
	})
	if err == nil {
		for _, alias := range aliasResp.Aliases {
			// NOTE: Aliases point to a published version. You may need to publish a new version
			// for the new Lambda and then create aliases pointing to that version.
			_, err = clientLamb.CreateAlias(ctx, &lambda.CreateAliasInput{
				FunctionName:    aws.String(functionNameNew),
				Name:            alias.Name,
				FunctionVersion: alias.FunctionVersion,
				Description:     alias.Description,
			})
			if err != nil {
				return fmt.Errorf("failed to create alias on new Lambda function: %v", err)
			}
		}
	}

	return nil
}

func (app *applicationMain) upgradeLambda(lambdaFunctionName string) error {
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		return fmt.Errorf("failed to create Lambda connection:\n%v", err)
	}

	newRuntime := types.RuntimePython313

	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(lambdaFunctionName),
		Runtime:      newRuntime,
	}

	_, err = clientLamb.UpdateFunctionConfiguration(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to udpate Lambda function\n%v", err)
	}

	return nil
}

func (app *applicationMain) listAllLambdaFunctions() (LambdaItems [][]string, err error) {
	ctx := context.Background()
	clientLamb, err := app.createLambdaClient()
	if err != nil {
		return LambdaItems, fmt.Errorf("failed to create Lambda connection:\n%v", err)
	}

	//create paginator
	paginator := lambda.NewListFunctionsPaginator(clientLamb, &lambda.ListFunctionsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return LambdaItems, fmt.Errorf("failed to retrieve next page of functions:\n%v", err)
		}

		for _, fn := range output.Functions {
			LambdaItems = append(LambdaItems, []string{aws.ToString(fn.FunctionName), string(fn.Runtime)})
		}
	}

	return LambdaItems, nil
}
