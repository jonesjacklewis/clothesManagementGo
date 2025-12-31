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

	// this shouldn't ever be possible, but as a guard. As a result, will not unit test.
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
			// this shouldn't ever be possible, but as a guard. As a result, will not unit test.
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

func (d *DynamoDBClothingRepository) GetById(id string) (domain.Clothing, error) {

	getItemInput := &dynamodb.GetItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	}

	getItemOutput, err := d.client.GetItem(context.TODO(), getItemInput)

	if err != nil {
		return domain.Clothing{}, fmt.Errorf("Failed to GetItem for id %s %v", id, err)
	}

	var item domain.Clothing

	if len(getItemOutput.Item) == 0 {
		return domain.Clothing{}, fmt.Errorf("No item found for id %s", id)
	}

	attributevalue.UnmarshalMap(getItemOutput.Item, &item)

	return item, nil
}

func (d *DynamoDBClothingRepository) Update(clothing domain.Clothing) (domain.Clothing, error) {
	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	if clothing.Id == "" {
		return domain.Clothing{}, fmt.Errorf("cannot update clothing without ID")
	}

	item, err := attributevalue.MarshalMap(clothing)

	// this shouldn't ever be possible, but as a guard. As a result, will not unit test.
	if err != nil {
		return domain.Clothing{}, fmt.Errorf("failed to marshal clothing item for DynamoDB: %w", err)
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	}

	_, err = d.client.PutItem(context.TODO(), putItemInput)
	if err != nil {
		return domain.Clothing{}, fmt.Errorf("failed to update item into DynamoDB: %w", err)
	}

	return clothing, nil
}

func (d *DynamoDBClothingRepository) Delete(id string) error {

	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("ID cannot be empty or whitespace")
	}

	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	}

	_, err := d.client.DeleteItem(context.TODO(), deleteItemInput)

	if err != nil {
		return fmt.Errorf("Failed to DeleteItem for id %s %v", id, err)
	}

	return nil
}

func (d *DynamoDBClothingRepository) Exists(id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("ID cannot be empty or whitespace")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{Value: id},
		},
		// Only fetch the key back
		ProjectionExpression: aws.String("Id"),
	}

	out, err := d.client.GetItem(context.TODO(), input)
	if err != nil {
		return false, fmt.Errorf("failed to check existence for id %s: %w", id, err)
	}

	return len(out.Item) != 0, nil
}
