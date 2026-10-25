// Package user_service provides shared user lookups and creation.
package user_service

import (
	"errors"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/database"
	"gorm.io/gorm"
)

// FindByEmail returns nil, nil when no user has the email.
func FindByEmail(email string) (*domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return findByEmail(db, email)
}

func findByEmail(db *gorm.DB, email string) (*domaine.User, error) {
	email = normalizeEmail(email)

	var user domaine.User
	err := db.Where("LOWER(email) = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID returns nil, nil when no user has the id.
func FindByID(id uint) (*domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return findByID(db, id)
}

func findByID(db *gorm.DB, id uint) (*domaine.User, error) {
	var user domaine.User
	err := db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new active user. If the email already exists, the existing
// user is returned instead of failing on the unique index.
func Create(email string) (*domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return create(db, email)
}

func create(db *gorm.DB, email string) (*domaine.User, error) {
	user := domaine.User{
		Email:    normalizeEmail(email),
		IsActive: true,
	}
	if err := db.Create(&user).Error; err != nil {
		existing, findErr := findByEmail(db, email)
		if findErr == nil && existing != nil {
			return existing, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindOrCreateByEmail returns an existing user or creates an active one.
func FindOrCreateByEmail(email string) (*domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return findOrCreateByEmail(db, email)
}

func findOrCreateByEmail(db *gorm.DB, email string) (*domaine.User, error) {
	user, err := findByEmail(db, email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}
	return create(db, email)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
