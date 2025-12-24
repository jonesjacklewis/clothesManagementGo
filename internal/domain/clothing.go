package domain

import (
	"errors"
	"strings"
)

type Clothing struct {
	Id           string `json:"id"`
	ClothingType string `json:"clothingType"`
	Description  string `json:"description"`
	Brand        string `json:"brand"`
	Store        string `json:"store"`
	ImageUrl     string `json:"imageUrl"`
	Price        Pence  `json:"pricePence"`
	Size         string `json:"size"`
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
