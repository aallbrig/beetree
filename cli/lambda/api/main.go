// Command api is the BeeTreeCraft API Lambda (REST + MCP behind API Gateway).
package main

import (
	"context"
	"log"
	"os"

	"github.com/aallbrig/beetree-cli/internal/api"
	"github.com/aallbrig/beetree-cli/internal/lambdahttp"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// version is stamped at build time with -ldflags.
var version = "dev"

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing required environment variable %s", name)
	}
	return v
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	srv := &api.Server{
		Store: &api.DynamoStore{DB: dynamodb.NewFromConfig(cfg), Table: mustEnv("TABLE_NAME")},
		Auth: &api.CognitoVerifier{
			Region:   mustEnv("AWS_REGION"),
			PoolID:   mustEnv("USER_POOL_ID"),
			ClientID: mustEnv("USER_POOL_CLIENT_ID"),
		},
		Version: version,
	}
	lambda.Start(lambdahttp.Adapt(srv.Handler()))
}
