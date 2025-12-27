package repository

import (
	"clothes_management/internal/domain"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/google/uuid"
)

type DynamoDBClothingRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBClothingRepository(client *dynamodb.Client, tableName string) (*DynamoDBClothingRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("client should not be nil")
	}

	if strings.TrimSpace(tableName) == "" {
		return nil, fmt.Errorf("tableName should not be empty or whitespace")
	}

	return &DynamoDBClothingRepository{
		client:    client,
		tableName: tableName,
	}, nil
}

func (d *DynamoDBClothingRepository) Save(clothing domain.Clothing) (domain.Clothing, error) {
	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	if clothing.Id == "" {
		clothing.Id = uuid.New().String()
	}

	item, err := attributevalue.MarshalMap(clothing)

	if err != nil {
		return domain.Clothing{}, fmt.Errorf("failed to marshal clothing item for DynamoDB: %w", err)
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	}

	_, err = d.client.PutItem(context.TODO(), putItemInput)
	if err != nil {
		return domain.Clothing{}, fmt.Errorf("failed to put item into DynamoDB: %w", err)
	}

	return clothing, nil

}

func (d *DynamoDBClothingRepository) GetAll() ([]domain.Clothing, error) {
	var allClothes []domain.Clothing
	var lastEvaluatedKey map[string]types.AttributeValue

	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	}

	for {
		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}

		output, err := d.client.Scan(context.TODO(), scanInput)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DynamoDB table '%s': %w", d.tableName, err)
		}

		var clothesPage []domain.Clothing
		err = attributevalue.UnmarshalListOfMaps(output.Items, &clothesPage)

		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal DynamoDB items from scan result: %w", err)
		}

		allClothes = append(allClothes, clothesPage...)

		if output.LastEvaluatedKey == nil {
			break
		}

		lastEvaluatedKey = output.LastEvaluatedKey
	}

	return allClothes, nil
}
