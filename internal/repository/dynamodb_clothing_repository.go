package repository

import (
	"clothes_management/internal/domain"
	"context"
	"errors"
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

func (d *DynamoDBClothingRepository) Save(userId string, clothing domain.Clothing) (domain.Clothing, error) {

	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	if clothing.Id == "" {
		clothing.Id = uuid.New().String()
	}

	clothing.UserId = userId

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

func (d *DynamoDBClothingRepository) GetAll(userId string) ([]domain.Clothing, error) {

	if strings.TrimSpace(userId) == "" {
		return []domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	var allClothes []domain.Clothing
	var lastEvaluatedKey map[string]types.AttributeValue

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		KeyConditionExpression: aws.String("UserId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userId},
		},
	}

	for {
		if lastEvaluatedKey != nil {
			queryInput.ExclusiveStartKey = lastEvaluatedKey
		}

		output, err := d.client.Query(context.TODO(), queryInput)
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

func (d *DynamoDBClothingRepository) GetById(userId, id string) (domain.Clothing, error) {

	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	if strings.TrimSpace(id) == "" {
		return domain.Clothing{}, errors.New("ID must not be empty or whitespace")
	}

	getItemInput := &dynamodb.GetItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"UserId": &types.AttributeValueMemberS{Value: userId},
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

func (d *DynamoDBClothingRepository) Update(userId string, clothing domain.Clothing) (domain.Clothing, error) {

	if strings.TrimSpace(userId) == "" {
		return domain.Clothing{}, errors.New("User ID must not be empty or whitespace")
	}

	if err := clothing.Validate(); err != nil {
		return domain.Clothing{}, err
	}

	if clothing.Id == "" {
		return domain.Clothing{}, fmt.Errorf("cannot update clothing without ID")
	}

	if clothing.UserId == "" {
		clothing.UserId = userId
	}

	if clothing.UserId != userId {
		return domain.Clothing{}, fmt.Errorf("Mismatch of user ID")
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

func (d *DynamoDBClothingRepository) Delete(userId, id string) error {

	if strings.TrimSpace(userId) == "" {
		return errors.New("User ID must not be empty or whitespace")
	}

	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("ID cannot be empty or whitespace")
	}

	exists, err := d.Exists(userId, id)

	if err != nil {
		return fmt.Errorf("Failed to DeleteItem for id %s %v", id, err)
	}

	if !exists {
		return fmt.Errorf("Item with id %s does not exist", id)
	}

	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"Id": &types.AttributeValueMemberS{
				Value: id,
			},
			"UserId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	}

	_, err = d.client.DeleteItem(context.TODO(), deleteItemInput)

	if err != nil {
		return fmt.Errorf("Failed to DeleteItem for id %s %v", id, err)
	}

	return nil
}

func (d *DynamoDBClothingRepository) Exists(userId, id string) (bool, error) {

	if strings.TrimSpace(userId) == "" {
		return false, errors.New("User ID must not be empty or whitespace")
	}

	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("ID cannot be empty or whitespace")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"Id":     &types.AttributeValueMemberS{Value: id},
			"UserId": &types.AttributeValueMemberS{Value: userId},
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
