package repository

import (
	"clothes_management/internal/domain"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials" // Corrected import
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
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

func setupLocalStackDynamoDBClient(t *testing.T, shouldClear bool) *dynamodb.Client {
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

	if shouldClear {
		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	}

	return client
}

func TestNewDynamoDBClothingRepository(t *testing.T) {

	t.Run("Given client is nil, should error", func(t *testing.T) {
		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		repo, err := NewDynamoDBClothingRepository(nil, dynamoTableName)

		if err == nil {
			t.Errorf("Expected to get an error, but didn't")
		}

		if repo != nil {
			t.Errorf("Expected repo to be nil")
		}

	})

	t.Run("Given tableName is empty, should error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

		dynamoTableName := ""

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err == nil {
			t.Errorf("Expected to get an error, but didn't")
		}

		if repo != nil {
			t.Errorf("Expected repo to be nil")
		}
	})

	t.Run("Given the constructor is called, should create a new instance", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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
		client := setupLocalStackDynamoDBClient(t, true)

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

	t.Run("Given table does not exist, should return an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		dynamoTableName = dynamoTableName + uuid.New().String()

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

		_, err = repo.Save(clothingItem)

		if err == nil {
			t.Error("Expected to error")
		}

		expectedMessage := "failed to put item into DynamoDB"

		if !strings.Contains(err.Error(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("Given valid clothing, should not error and should have an ID", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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
		client := setupLocalStackDynamoDBClient(t, true)

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

	t.Run("Given table doesn't exist, should error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		dynamoTableName = dynamoTableName + uuid.New().String()

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

		if err == nil {
			t.Errorf("Expected error when saving")
		}

		_, err = repo.GetAll()

		if err == nil {
			t.Errorf("Expected error on GetAll")
		}

		expectedMessage := "failed to scan DynamoDB table"

		if !strings.Contains(err.Error(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When there is 1 clothing item, GetAll should return an list with len = 1 and no errors", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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
		client := setupLocalStackDynamoDBClient(t, true)

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

func TestDynamoGetById(t *testing.T) {

	t.Run("When table doesn't exist, should return an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, false)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		dynamoTableName = dynamoTableName + uuid.New().String()

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		_, err = repo.GetById("dummy-id-123")

		expectedMessage := "Failed to GetItem for id"

		if err == nil {
			t.Fatalf("Expected to get an err")
		}

		if !strings.Contains(err.Error(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When the item doesn't exist GetById should return an error", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		_, err := repo.GetById("dummy-id-123")

		if err == nil {
			t.Errorf("Expected an error")
		}

		expectedMessage := "No item exists for id dummy-id-123"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When the item exists, GetById should return the item", func(t *testing.T) {
		repo := NewInMemoryClothingRepository()

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		if repo.items == nil {
			t.Fatal("repo.items should not be null")
		}

		if len(repo.items) != 0 {
			t.Errorf("Expected repo.items.length = 0, got %d", len(repo.items))
		}

		repo.mu.Lock()

		dummyId := "dummy-id-123"

		repo.items[dummyId] = domain.Clothing{
			Id:           dummyId,
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		repo.mu.Unlock()

		item, err := repo.GetById(dummyId)

		if err != nil {
			t.Errorf("Expected not error, got %v", err)
		}

		if item.Id != dummyId {
			t.Errorf("Expected item.Id = %s, got %s", dummyId, item.Id)
		}
	})
}

func TestDynamoUpdate(t *testing.T) {
	t.Run("When the item isn't valid should return an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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

		dummyId := "dummy-id-123"

		item := domain.Clothing{
			Id:           dummyId,
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        -2000,
		}

		_, err = repo.Update(item)

		if err == nil {
			t.Errorf("Expected an error")
		}

		expectedMessage := "Clothing Price must be greater than or equal to 0"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When the item doesn't have an id should return an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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

		_, err = repo.Update(item)

		if err == nil {
			t.Errorf("Expected an error")
		}

		expectedMessage := "cannot update clothing without ID"

		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When table doesn't exist, should return an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, false)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		dynamoTableName = dynamoTableName + uuid.New().String()

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)

		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}

		if repo == nil {
			t.Fatal("repo should not be null")
		}

		item := domain.Clothing{
			Id:           "dummy-id-123",
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        2000,
		}

		_, err = repo.Update(item)

		expectedMessage := "failed to update item into DynamoDB:"

		if err == nil {
			t.Fatalf("Expected to get an err")
		}

		if !strings.Contains(err.Error(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}

	})

	t.Run("When item is valid, should update item", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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

		originalPrice := domain.Pence(2000)

		item := domain.Clothing{
			ClothingType: "Jumper",
			Description:  "This Jumper",
			Store:        "This Store",
			Size:         "L",
			Brand:        "XYZ",
			Price:        originalPrice,
		}

		item, err = repo.Save(item)

		if err != nil {
			t.Fatalf("Expected not to err, got %v", err)
		}

		newPrice := domain.Pence(1500)

		item.Price = newPrice

		item, err = repo.Update(item)

		if item.Price == originalPrice {
			t.Errorf("Expected %d got %d", newPrice, originalPrice)
		}

		item, err = repo.GetById(item.Id)

		if err != nil {
			t.Fatalf("Expected not to err, got %v", err)
		}

		if item.Price == originalPrice {
			t.Errorf("Expected %d got %d", newPrice, originalPrice)
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})

	})

}

func TestDynamoExists(t *testing.T) {
	t.Run("Given empty or whitespace ID, should return false and an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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

		id := ""

		exists, err := repo.Exists(id)

		if exists {
			t.Error("Expected exists = false")
		}
		if err == nil {
			t.Fatal("Expected to get an error")
		}

		expectedMessage := "ID cannot be empty or whitespace"
		if err.Error() != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, err.Error())
		}
	})

	t.Run("Given table doesn't exist, should return false and an error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, false)

		dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
		if dynamoTableName == "" {
			t.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env_test or your shell.")
		}

		dynamoTableName = dynamoTableName + uuid.New().String()

		repo, err := NewDynamoDBClothingRepository(client, dynamoTableName)
		if err != nil {
			t.Fatalf("Expected no err on NewDynamoDBClothingRepository, got %v", err)
		}
		if repo == nil {
			t.Fatal("repo should not be null")
		}

		exists, err := repo.Exists("dummy-id-123")

		if exists {
			t.Error("Expected exists = false")
		}
		if err == nil {
			t.Fatal("Expected an error")
		}

		expectedMessage := "failed to check existence for id"
		if !strings.Contains(err.Error(), expectedMessage) {
			t.Errorf("Expected error to contain %q, got %q", expectedMessage, err.Error())
		}
	})

	t.Run("Given item does not exist, should return false and no error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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

		exists, err := repo.Exists("does-not-exist")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if exists {
			t.Error("Expected exists = false")
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	})

	t.Run("Given item exists, should return true and no error", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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

		saved, err := repo.Save(item)
		if err != nil {
			t.Fatalf("Expected no error saving item, got %v", err)
		}
		if saved.Id == "" {
			t.Fatal("Expected saved item to have an ID")
		}

		exists, err := repo.Exists(saved.Id)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !exists {
			t.Error("Expected exists = true")
		}

		t.Cleanup(func() {
			clearDynamoDBTable(t, client, dynamoTableName)
		})
	})
}

func TestDynamoConcurrentSaves(t *testing.T) {
	t.Run("Given multiple goroutines concurrently save items, all items should be saved correctly", func(t *testing.T) {
		client := setupLocalStackDynamoDBClient(t, true)

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
		client := setupLocalStackDynamoDBClient(t, true)

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
