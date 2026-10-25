// Package transaction_service creates and loads credit purchase transactions.
package transaction_service

import (
	"errors"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create starts a pending transaction for the given user and optional vehicle.
func Create(userID uint, cacheRecordID *uint, credits, amount uint) (*domaine.Transaction, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return create(db, userID, cacheRecordID, credits, amount)
}

func create(db *gorm.DB, userID uint, cacheRecordID *uint, credits, amount uint) (*domaine.Transaction, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	transaction := &domaine.Transaction{
		Token:         token.String(),
		Credits:       credits,
		Amount:        amount,
		Status:        domaine.TransactionPending,
		PaymentMethod: domaine.PaymentMethodPayPalCheckout,
		UserID:        userID,
		CacheRecordID: cacheRecordID,
	}
	if err := db.Create(transaction).Error; err != nil {
		return nil, err
	}
	return transaction, nil
}

// FindByToken returns nil, nil when the token is malformed or no transaction
// exists. The user and optional vehicle are preloaded for the payment page.
func FindByToken(token string) (*domaine.Transaction, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return findByToken(db, token)
}

func findByToken(db *gorm.DB, token string) (*domaine.Transaction, error) {
	if len(token) != 36 {
		return nil, nil
	}
	if _, err := uuid.Parse(token); err != nil {
		return nil, nil
	}

	var transaction domaine.Transaction
	err := db.Preload("User").
		Preload("CacheRecord").
		Where("token = ?", token).
		First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}
