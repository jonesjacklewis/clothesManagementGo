package domain

import (
	"errors"
	"strings"
)

type Clothing struct {
	Id           string `json:"id" dynamodbav:"Id"`
	UserId       string `json:"userId" dynamodbav:"UserId"`
	ClothingType string `json:"clothingType" dynamodbav:"ClothingType"`
	Description  string `json:"description" dynamodbav:"Description"`
	Brand        string `json:"brand" dynamodbav:"Brand"`
	Store        string `json:"store" dynamodbav:"Store"`
	ImageUrl     string `json:"imageUrl" dynamodbav:"ImageUrl"`
	Price        Pence  `json:"pricePence" dynamodbav:"PricePence"`
	Size         string `json:"size" dynamodbav:"Size"`
}

func (c Clothing) Validate() error {
	if int(c.Price) < 0 {
		return errors.New("Clothing Price must be greater than or equal to 0")
	}

	if strings.TrimSpace(c.ClothingType) == "" {
		return errors.New("Clothing Type must not be empty")
	}

	if strings.TrimSpace(c.Description) == "" {
		return errors.New("Clothing Description must not be empty")
	}

	if strings.TrimSpace(c.Brand) == "" {
		return errors.New("Clothing Brand must not be empty")
	}

	if strings.TrimSpace(c.Store) == "" {
		return errors.New("Clothing Store must not be empty")
	}

	if strings.TrimSpace(c.Size) == "" {
		return errors.New("Clothing Size must not be empty")
	}

	return nil
}
