package repository

import (
	"clothes_management/internal/domain"
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials" // Corrected import
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env_test"); err != nil {
		fmt.Printf("INFO (test): No .env_test file found or error loading: %v. Relying on system environment variables.\n", err)
	} else {
		fmt.Println("INFO (test): .env_test file loaded for tests.")
	}

	// Run all tests in the package
	os.Exit(m.Run())
}

func clearDynamoDBTable(t *testing.T, client *dynamodb.Client, tableName string) {
	t.Helper() // Marks this as a test helper function

	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		scanOutput, err := client.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:            aws.String(tableName),
			ProjectionExpression: aws.String("Id"),
			ExclusiveStartKey:    lastEvaluatedKey,
		})
		if err != nil {
			t.Fatalf("Failed to scan table for cleanup: %v", err)
		}

		if len(scanOutput.Items) == 0 {
			break
		}

		writeRequests := make([]types.WriteRequest, 0, len(scanOutput.Items))
		for _, item := range scanOutput.Items {
			writeRequests = append(writeRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: item,
				},
			})
		}
		for i := 0; i < len(writeRequests); i += 25 {
			end := i + 25
			if end > len(writeRequests) {
				end = len(writeRequests)
			}
			batch := writeRequests[i:end]

			_, err = client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					tableName: batch,
				},
			})
			if err != nil {
				t.Fatalf("Failed to batch delete items for cleanup: %v", err)
			}
		}

		lastEvaluatedKey = scanOutput.LastEvaluatedKey
		if lastEvaluatedKey == nil {
			break
		}
	}
	t.Logf("Cleaned up DynamoDB table '%s'", tableName)
}

func setupLocalStackDynamoDBClient(t *testing.T) *dynamodb.Client {
	accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyId == "" {
		t.Fatal("ERROR: AWS_ACCESS_KEY_ID environment variable not set. Please set it in .env_test or your shell.")
	}
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		t.Fatal("ERROR: AWS_SECRET_ACCESS_KEY environment variable not set. Please set it in .env_test or your shell.")
	}
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		t.Fatal("ERROR: AWS_REGION environment variable not set. Please set it in .env_test or your shell.")
	}
	dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if dynamoTableName == "" {
		t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
	}
	baseEndoint := os.Getenv("BASE_ENDPOINT")
	if baseEndoint == "" {
		t.Fatal("ERROR: BASE_ENDPOINT environment variable not set. Please set it in .env_test or your shell.")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
	)
	if err != nil {
		t.Fatalf("Failed to load AWS SDK config for LocalStack: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(baseEndoint)
	})

	t.Cleanup(func() {
		clearDynamoDBTable(t, client, dynamoTableName)
	})

	return client
}

func TestNewDynamoDBClothingRepository(t *testing.T) {
	t.Run("Given the constructor is called, should create a new instance", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.client == nil {
			t.Fatal("repo.client should not be null")
		}

		if repo.tableName == "" {
			t.Fatal("repo.tableName should not be empty")
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	})
}

func TestDynamoSave(t *testing.T) {
	t.Run("Given invalid clothing, should return error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		clothingItem := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        -2000,
		}

		item, err := repo.Save(clothingItem)

		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if item.Id != "" {
			t.Errorf("Expected item not to have id, got %s", item.Id)
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})

	})

	t.Run("Given valid clothing, should not error and should have an ID", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		clothingItem := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		item, err := repo.Save(clothingItem)

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if item.Id == "" {
			t.Error("Expected item to be populated")
		}

		if item.Description != clothingItem.Description {
			t.Errorf("Expected %s got %s", clothingItem.Description, item.Description)
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})

	})
}

func TestDynamoGetAll(t *testing.T) {
	t.Run("When there are no clothing items, GetAll should return an empty list and no errors", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if len(items) > 0 {
			t.Errorf("Expected no items, got %d", len(items))
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	})

	t.Run("When there is 1 clothing item, GetAll should return an list with len = 1 and no errors", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		item := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		_, err = repo.Save(item)

		if err != nil {
			t.Errorf("Expected no error when saving %s, got %v", item.Description, err)
		}

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if len(items) != 1 {
			t.Fatalf("Expected 1 item, got %d", len(items))
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})

	})

	t.Run("When there is more than 1 clothing item, GetAll should return an list with len > 1 and no errors", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		item := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		_, err = repo.Save(item)

		if err != nil {
			t.Errorf("Expected no error when saving %s, got %v", item.Description, err)
		}

		item = domain.Clothing{
			ClothingType: "Sweater",
			Description:  "This Sweater",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		_, err = repo.Save(item)

		if err != nil {
			t.Errorf("Expected no error when saving %s, got %v", item.Description, err)
		}

		items, err := repo.GetAll()

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if len(items) < 2 {
			t.Errorf("Expected two items, got %d", len(items))
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	})
}

func TestDynamoConcurrentSaves(t *testing.T) {
	t.Run("Given multiple goroutines concurrently save items, all items should be saved correctly", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		numConcurrentSaves := 100
		var wg sync.WaitGroup

		for i := 0; i < numConcurrentSaves; i++ {
			wg.Add(1)

			go func(index int) {
				defer wg.Done()

				clothingItem := domain.Clothing{
					ClothingType: fmt.Sprintf("Jumper %d", index),
					Description:  fmt.Sprintf("Concurrent Jumper %d", index),
					Brand:        "ConcurrentBrand",
					Store:        "ConcurrentStore",
					Size:         "M",
					Price:        domain.Pence(1000 + index),
				}

				_, err := repo.Save(clothingItem)

				if err != nil {
					t.Errorf("Goroutine %d: Failed to save item: %v", index, err)
				}

			}(i)

		}

		wg.Wait()

		allSavedItems, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed to retrieve all items after concurrent saves: %v", err)
		}

		if len(allSavedItems) != numConcurrentSaves {
			t.Errorf("Expected %d items after concurrent saves, got %d", numConcurrentSaves, len(allSavedItems))
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})

	})
}

func TestDynamoConcurrentGetAll(t *testing.T) {
	t.Run("Given multiple goroutines concurrently retrieve items, all retrievals should be consistent and error-free", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		numInitialItems := 10
		for i := 0; i < numInitialItems; i++ {
			clothingItem := domain.Clothing{
				ClothingType: fmt.Sprintf("Initial Jumper %d", i),
				Description:  fmt.Sprintf("Initial Description %d", i),
				Brand:        "InitialBrand",
				Store:        "InitialStore",
				Size:         "S",
				Price:        domain.Pence(500 + i),
			}
			_, err := repo.Save(clothingItem)
			if err != nil {
				t.Fatalf("Failed to setup initial item for concurrent GetAll test: %v", err)
			}
		}

		numConcurrentReads := 500
		var wg sync.WaitGroup
		for i := 0; i < numConcurrentReads; i++ {
			wg.Add(1)
			go func(readIndex int) {
				defer wg.Done()

				items, err := repo.GetAll()
				if err != nil {
					t.Errorf("Goroutine %d: Failed to retrieve items: %v", readIndex, err)
					return
				}

				if len(items) != numInitialItems {
					t.Errorf("Goroutine %d: Expected %d items, got %d", readIndex, numInitialItems, len(items))
				}
			}(i)
		}

		wg.Wait()

		finalItems, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Failed final GetAll check: %v", err)
		}
		if len(finalItems) != numInitialItems {
			t.Errorf("Final check: Expected %d items, got %d", numInitialItems, len(finalItems))
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	})
}
