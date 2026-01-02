package repository

import (
	"clothes_management/internal/domain"
)

type ClothingRepository interface {
	Save(userId string, clothing domain.Clothing) (domain.Clothing, error)
	GetAll(userId string) ([]domain.Clothing, error)
	GetById(userId, id string) (domain.Clothing, error)
	Update(userId string, clothing domain.Clothing) (domain.Clothing, error)
	Delete(userId, id string) error
	Exists(userId, id string) (bool, error)
}
