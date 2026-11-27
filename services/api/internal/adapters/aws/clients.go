package aws

import (
	"context"
	"errors"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// LoadConfig loads the default AWS configuration (environment, Lambda role). It is used only by
// the wiring layer.
func LoadConfig(ctx context.Context) (awssdk.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return awssdk.Config{}, fmt.Errorf("aws config: %w", err)
	}
	return cfg, nil
}

// NewDynamoDBClient returns a client for the configured region, or for a local endpoint such as
// DynamoDB Local when endpoint is not empty (tests and the development server only).
func NewDynamoDBClient(cfg awssdk.Config, endpoint string) *dynamodb.Client {
	if endpoint == "" {
		return dynamodb.NewFromConfig(cfg)
	}
	return dynamodb.New(dynamodb.Options{
		Region:       "local",
		BaseEndpoint: awssdk.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider("local", "local", ""),
	})
}

// SecretString reads a secret's string value from Secrets Manager. The OAuth App client secret
// is read this way at start-up (D-30); it is never logged or persisted elsewhere.
func SecretString(ctx context.Context, client *secretsmanager.Client, secretID string) (string, error) {
	if secretID == "" {
		return "", errors.New("secrets manager: secret ID is required")
	}
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: awssdk.String(secretID)})
	if err != nil {
		return "", fmt.Errorf("secrets manager: reading %q: %w", secretID, err)
	}
	if out.SecretString == nil || *out.SecretString == "" {
		return "", fmt.Errorf("secrets manager: %q has no string value", secretID)
	}
	return *out.SecretString, nil
}
