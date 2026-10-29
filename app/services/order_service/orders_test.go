package order_service

import (
	"errors"
	"path/filepath"
	"testing"

	"carf-chorf.com/app/app/domaine"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := filepath.Join(t.TempDir(), "orders.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.User{}, &domaine.CacheRecord{}, &domaine.Order{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCreateOrderForCredit(t *testing.T) {
	db := openTestDB(t)

	user := &domaine.User{Email: "user@example.com", IsActive: true, Credits: 3}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	record := &domaine.CacheRecord{
		Token:    uuid.NewString(),
		Vin:      "1HGCM82633A004352",
		Provider: domaine.CarbpiProvider,
	}
	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create record: %v", err)
	}

	order, err := createOrderForCredit(db, user.ID, record, uuid.NewString())
	if err != nil {
		t.Fatalf("createOrderForCredit() error = %v", err)
	}
	if order.ID == 0 || order.AccessToken == "" {
		t.Fatal("order not created")
	}
	if order.TransactionID != nil {
		t.Fatalf("transaction id = %v, want nil", order.TransactionID)
	}
	if order.CacheRecordID != record.ID || order.Vin != record.Vin {
		t.Fatalf("unexpected order: %+v", order)
	}

	var updated domaine.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if updated.Credits != 2 {
		t.Fatalf("credits = %d, want 2", updated.Credits)
	}

	orders, err := listByUser(db, user.ID)
	if err != nil {
		t.Fatalf("listByUser() error = %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("orders = %d, want 1", len(orders))
	}
}

func TestCreateOrderForCreditInsufficientCredits(t *testing.T) {
	db := openTestDB(t)

	user := &domaine.User{Email: "user@example.com", IsActive: true, Credits: 0}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	record := &domaine.CacheRecord{
		Token:    uuid.NewString(),
		Vin:      "1HGCM82633A004352",
		Provider: domaine.CarbpiProvider,
	}
	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create record: %v", err)
	}

	_, err := createOrderForCredit(db, user.ID, record, uuid.NewString())
	if !errors.Is(err, ErrNotEnoughCredits) {
		t.Fatalf("error = %v, want ErrNotEnoughCredits", err)
	}

	var count int64
	if err := db.Model(&domaine.Order{}).Count(&count).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if count != 0 {
		t.Fatalf("order count = %d, want 0", count)
	}
}
