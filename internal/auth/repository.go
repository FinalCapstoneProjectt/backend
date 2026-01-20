package auth

import (
	"backend/internal/domain"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	Update(user *domain.User) error
	IncrementFailedLogins(userID uint) error
	ResetFailedLogins(userID uint) error
	UpdateLastLogin(userID uint) error
	LockAccount(userID uint, until time.Time) error
	IsAccountLocked(userID uint) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *repository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).
		Preload("University").
		Preload("Department").
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ?", id).
		Preload("University").
		Preload("Department").
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *repository) IncrementFailedLogins(userID uint) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + ?", 1)).
		Error
}

func (r *repository) ResetFailedLogins(userID uint) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"account_locked_until":  nil,
		}).
		Error
}

func (r *repository) UpdateLastLogin(userID uint) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("last_login_at", time.Now()).
		Error
}

func (r *repository) LockAccount(userID uint, until time.Time) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("account_locked_until", until).
		Error
}

func (r *repository) IsAccountLocked(userID uint) (bool, error) {
	var user domain.User
	err := r.db.Select("account_locked_until").First(&user, userID).Error
	if err != nil {
		return false, err
	}

	if user.AccountLockedUntil != nil && time.Now().Before(*user.AccountLockedUntil) {
		return true, nil
	}
	return false, nil
}
