package auth

import "gorm.io/gorm"

type Repository interface {
	// Interface methods go here
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
