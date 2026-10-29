// Package order_service creates and lists purchased report orders.
package order_service

import (
	"context"
	"errors"
	"fmt"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/database"
	"carf-chorf.com/app/app/services/discord_service"
	"carf-chorf.com/app/app/services/report_service"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotEnoughCredits = errors.New("not enough credits")

// BuyReportWithCredit prepares and saves a report, then consumes one credit and
// creates an order in a single database transaction.
func BuyReportWithCredit(ctx context.Context, userID uint, record *domaine.CacheRecord) (*domaine.Order, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return buyReportWithCredit(ctx, db, userID, record)
}

func buyReportWithCredit(ctx context.Context, db *gorm.DB, userID uint, record *domaine.CacheRecord) (*domaine.Order, error) {
	if record == nil {
		return nil, errors.New("order_service: vehicle record is required")
	}

	var credits uint
	if err := db.Model(&domaine.User{}).Where("id = ?", userID).Pluck("credits", &credits).Error; err != nil {
		return nil, err
	}
	if credits < 1 {
		return nil, ErrNotEnoughCredits
	}

	accessToken, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	token := accessToken.String()

	html, err := report_service.Prepare(ctx, record.Vin)
	if err != nil {
		sendSOS(record.Vin, userID, err)
		return nil, fmt.Errorf("order_service: prepare report: %w", err)
	}
	if err := report_service.Save(token, html); err != nil {
		_ = report_service.Delete(token)
		return nil, fmt.Errorf("order_service: save report: %w", err)
	}

	order, err := createOrderForCredit(db, userID, record, token)
	if err != nil {
		_ = report_service.Delete(token)
		return nil, err
	}
	return order, nil
}

func createOrderForCredit(db *gorm.DB, userID uint, record *domaine.CacheRecord, accessToken string) (*domaine.Order, error) {
	order := domaine.Order{
		UserID:        userID,
		CacheRecordID: record.ID,
		Vin:           record.Vin,
		AccessToken:   accessToken,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&domaine.User{}).
			Where("id = ? AND credits >= 1", userID).
			UpdateColumn("credits", gorm.Expr("credits - 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotEnoughCredits
		}

		return tx.Create(&order).Error
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ListByUser returns a user's orders, newest first.
func ListByUser(userID uint) ([]domaine.Order, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	return listByUser(db, userID)
}

func listByUser(db *gorm.DB, userID uint) ([]domaine.Order, error) {
	var orders []domaine.Order
	err := db.Preload("CacheRecord").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func sendSOS(vin string, userID uint, err error) {
	message := fmt.Sprintf("🆘 **Report purchase failed**\nVIN: `%s`\nUser ID: `%d`\nError: %v", vin, userID, err)
	go func() {
		if notifyErr := discord_service.SendSOSNotification(message); notifyErr != nil {
			simple_logger.Errorf("order_service: send SOS: %v", notifyErr)
		}
	}()
}
