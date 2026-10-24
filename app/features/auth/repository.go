package auth

import (
	"errors"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/database"
	"gorm.io/gorm"
)

func findUserByEmail(email string) (*domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	var user domaine.User
	err = db.Where("LOWER(email) = ?", strings.ToLower(email)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func createUser(email string) (*domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	user := domaine.User{
		Email:    strings.ToLower(email),
		IsActive: true,
	}
	if err := db.Create(&user).Error; err != nil {
		existing, findErr := findUserByEmail(email)
		if findErr == nil && existing != nil {
			return existing, nil
		}
		return nil, err
	}

	return &user, nil
}
