package api

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// TestDynamoStoreContract runs the shared Store contract against a real table.
// It is skipped unless BEETREE_TEST_TABLE names a table that is safe to write
// to; the post-deploy smoke test points it at a throwaway stack.
func TestDynamoStoreContract(t *testing.T) {
	table := os.Getenv("BEETREE_TEST_TABLE")
	if table == "" {
		t.Skip("BEETREE_TEST_TABLE not set")
	}
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	runStoreContract(t, &DynamoStore{DB: dynamodb.NewFromConfig(cfg), Table: table})
}
