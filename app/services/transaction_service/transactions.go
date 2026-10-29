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

var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrTransactionNotPending  = errors.New("transaction is not pending")
	ErrTransactionPidMismatch = errors.New("transaction pid mismatch")
	ErrNotEnoughCredits       = errors.New("not enough credits")
)

// CompletePurchase marks a transaction paid, adds its credits, and creates a
// report order when the transaction has a vehicle attached. It is idempotent:
// repeated captures return the existing order without adding credits again.
func CompletePurchase(token, pid, reportAccessToken string) (*domaine.Transaction, *domaine.Order, bool, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, nil, false, err
	}
	return completePurchase(db, token, pid, reportAccessToken)
}

func completePurchase(db *gorm.DB, token, pid, reportAccessToken string) (*domaine.Transaction, *domaine.Order, bool, error) {
	var transaction domaine.Transaction
	var order domaine.Order
	orderCreated := false

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Preload("User").
			Preload("CacheRecord").
			Where("token = ?", token).
			First(&transaction).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTransactionNotFound
			}
			return err
		}

		if pid != "" && transaction.Pid != "" && transaction.Pid != pid {
			return ErrTransactionPidMismatch
		}

		hasOrder := false
		if transaction.CacheRecordID != nil {
			orderErr := tx.Where("transaction_id = ?", transaction.ID).First(&order).Error
			switch {
			case orderErr == nil:
				hasOrder = true
			case errors.Is(orderErr, gorm.ErrRecordNotFound):
			default:
				return orderErr
			}
		}

		if transaction.Status != domaine.TransactionPaid {
			if transaction.Status != domaine.TransactionPending {
				return ErrTransactionNotPending
			}

			updates := map[string]any{"status": domaine.TransactionPaid}
			if pid != "" {
				updates["pid"] = pid
			}
			result := tx.Model(&domaine.Transaction{}).
				Where("id = ? AND status = ?", transaction.ID, domaine.TransactionPending).
				Updates(updates)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrTransactionNotPending
			}

			if err := tx.Model(&domaine.User{}).
				Where("id = ?", transaction.UserID).
				UpdateColumn("credits", gorm.Expr("credits + ?", transaction.Credits)).Error; err != nil {
				return err
			}
		}

		if transaction.CacheRecordID != nil && !hasOrder {
			if transaction.CacheRecord == nil {
				return errors.New("transaction vehicle record is missing")
			}

			result := tx.Model(&domaine.User{}).
				Where("id = ? AND credits >= 1", transaction.UserID).
				UpdateColumn("credits", gorm.Expr("credits - 1"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrNotEnoughCredits
			}

			accessToken := reportAccessToken
			if accessToken == "" {
				generated, err := uuid.NewRandom()
				if err != nil {
					return err
				}
				accessToken = generated.String()
			} else if _, err := uuid.Parse(accessToken); err != nil {
				return errors.New("invalid report access token")
			}

			transactionID := transaction.ID
			order = domaine.Order{
				UserID:        transaction.UserID,
				CacheRecordID: *transaction.CacheRecordID,
				TransactionID: &transactionID,
				Vin:           transaction.CacheRecord.Vin,
				AccessToken:   accessToken,
			}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
			orderCreated = true
		}

		return tx.Preload("User").
			Preload("CacheRecord").
			Where("id = ?", transaction.ID).
			First(&transaction).Error
	})
	if err != nil {
		return nil, nil, false, err
	}
	if transaction.CacheRecordID == nil {
		order = domaine.Order{}
	}
	return &transaction, &order, orderCreated, nil
}

// FindOrderByAccessToken returns nil, nil for malformed or unknown tokens.
func FindOrderByAccessToken(token string) (*domaine.Order, error) {
	if len(token) != 36 {
		return nil, nil
	}
	if _, err := uuid.Parse(token); err != nil {
		return nil, nil
	}

	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	var order domaine.Order
	err = db.Preload("User").
		Preload("CacheRecord").
		Where("access_token = ?", token).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// SetPid stores the PayPal order id on a pending transaction.
func SetPid(token, pid string) error {
	db, err := database.GetDB()
	if err != nil {
		return err
	}
	result := db.Model(&domaine.Transaction{}).
		Where("token = ? AND status = ?", token, domaine.TransactionPending).
		Update("pid", pid)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTransactionNotPending
	}
	return nil
}

// ListByUser returns a user's transactions, newest first.
func ListByUser(userID uint) ([]domaine.Transaction, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return listByUser(db, userID)
}

func listByUser(db *gorm.DB, userID uint) ([]domaine.Transaction, error) {
	var transactions []domaine.Transaction
	err := db.Preload("CacheRecord").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// ListPaidByUser returns a user's completed transactions, newest first.
func ListPaidByUser(userID uint) ([]domaine.Transaction, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return listPaidByUser(db, userID)
}

func listPaidByUser(db *gorm.DB, userID uint) ([]domaine.Transaction, error) {
	var transactions []domaine.Transaction
	err := db.Preload("CacheRecord").
		Where("user_id = ? AND status = ?", userID, domaine.TransactionPaid).
		Order("created_at DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
