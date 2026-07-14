package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type jwtSecret struct {
	Secret string `json:"secret"`
}

type mongoSecret struct {
	URI string `json:"uri"`
}

const parameterPath = "/gerek/app"

func loadAWS(ctx context.Context) (*Config, error) {

	awsCfg, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)
	secretsClient := secretsmanager.NewFromConfig(awsCfg)

	params, err := loadParameters(ctx, ssmClient)
	if err != nil {
		return nil, err
	}

	var jwt jwtSecret
	if err := loadSecret(ctx, secretsClient, params["jwt-secret-arn"], &jwt); err != nil {
		return nil, err
	}

	var mongo mongoSecret
	if err := loadSecret(ctx, secretsClient, params["mongo-uri-secret-arn"], &mongo); err != nil {
		return nil, err
	}

	return &Config{
		ServiceName:         getOrDefault(params, "service-name", "chat-service"),
		AppEnv:              getOrDefault(params, "env", "production"),
		Port:                getOrDefault(params, "server-port", "8081"),
		GRPCPort:            getOrDefault(params, "grpc-port", "9090"),
		JWTSecret:           jwt.Secret,
		MongoURI:            mongo.URI,
		MongoDB:             params["mongo-db-name"],
		UserServiceGRPCAddr: getOrDefault(params, "user-service-grpc-addr", "localhost:9090"),
		MetricsServerPort:   getOrDefault(params, "metrics-server-port", ":9100"),
		OTelCollectorAddr:   getOrDefault(params, "otel-collector-addr", "localhost:4317"),
	}, nil
}

func loadParameters(ctx context.Context, client *ssm.Client) (map[string]string, error) {
	params := make(map[string]string)

	var nextToken *string

	for {
		out, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String(parameterPath),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			NextToken:      nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, p := range out.Parameters {
			name := strings.TrimPrefix(aws.ToString(p.Name), parameterPath+"/")
			params[name] = aws.ToString(p.Value)
		}

		if out.NextToken == nil {
			break
		}

		nextToken = out.NextToken
	}

	return params, nil
}

func loadSecret(ctx context.Context, client *secretsmanager.Client, arn string, dst any) error {
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(aws.ToString(out.SecretString)), dst)
}

func getOrDefault(values map[string]string, key, def string) string {
	envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))

	if v := os.Getenv(envKey); v != "" {
		return v
	}

	if v, ok := values[key]; ok && v != "" {
		return v
	}

	return def
}
