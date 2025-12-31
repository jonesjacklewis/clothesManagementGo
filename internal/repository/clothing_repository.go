package repository

import (
	"clothes_management/internal/domain"
)

type ClothingRepository interface {
	Save(clothing domain.Clothing) (domain.Clothing, error)
	GetAll() ([]domain.Clothing, error)
	GetById(id string) (domain.Clothing, error)
	Update(clothing domain.Clothing) (domain.Clothing, error)
	Delete(id string) error
	Exists(id string) (bool, error)
}
