package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const ParameterPath = "/gig/app"

const (
	ParameterMongoURISecretArn             = "mongo-uri-secret-arn"
	ParameterMongoDBName                   = "mongo-db-name"
	ParameterSQSNotificationEventsQueueURL = "sqs-notification-events-queue-url"
	ParameterJWTSecretARN                  = "jwt-secret-arn"
)

type jwtSecret struct {
	Secret string `json:"secret"`
}

type mongoSecret struct {
	URI string `json:"uri"`
}

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

	if err := loadSecret(ctx, secretsClient, params[ParameterJWTSecretARN], &jwt); err != nil {
		return nil, err
	}

	var mongo mongoSecret

	if err := loadSecret(ctx, secretsClient, params[ParameterMongoURISecretArn], &mongo); err != nil {
		return nil, err
	}

	return &Config{
		ServiceName:         getEnv(EnvServiceName, DefaultServiceName),
		AppEnv:              EnvironmentProduction,
		Port:                getEnv(EnvServerPort, DefaultServerPort),
		JWTSecret:           jwt.Secret,
		MongoURI:            mongo.URI,
		MongoDB:             params[ParameterMongoDBName],
		UserServiceGRPCAddr: getEnv(EnvUserServiceGRPCAddr, DefaultUserServiceGRPCAddr),
		MetricsServerPort:   getEnv(EnvMetricsServerPort, DefaultMetricsServerPort),
		OTelCollectorAddr:   getEnv(EnvOTelCollectorAddr, DefaultOtelCollectorAddr),
	}, nil
}

func loadParameters(ctx context.Context, client *ssm.Client) (map[string]string, error) {
	names := []string{
		parameter(ParameterJWTSecretARN),
		parameter(ParameterMongoURISecretArn),
		parameter(ParameterMongoDBName),
		parameter(ParameterSQSNotificationEventsQueueURL),
	}

	out, err := client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return nil, fmt.Errorf("read parameter store: %w", err)
	}

	params := make(map[string]string)

	for _, p := range out.Parameters {
		key := strings.TrimPrefix(aws.ToString(p.Name), ParameterPath+"/")
		params[key] = aws.ToString(p.Value)
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

func parameter(name string) string {
	return ParameterPath + "/" + name
}
