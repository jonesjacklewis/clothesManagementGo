package repository

import (
	"clothes_management/internal/domain"
)

type ClothingRepository interface {
	Save(clothing domain.Clothing) (domain.Clothing, error)
	GetAll() ([]domain.Clothing, error)
}
