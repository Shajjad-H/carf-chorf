package transaction_service

import (
	"path/filepath"
	"testing"

	"carf-chorf.com/app/app/domaine"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateAndFindTransaction(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "transactions.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.User{}, &domaine.CacheRecord{}, &domaine.Transaction{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	user := &domaine.User{Email: "user@example.com", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	record := &domaine.CacheRecord{
		Token:         uuid.NewString(),
		Vin:           "1HGCM82633A004352",
		Provider:      domaine.CarbpiProvider,
		Vehicle:       "2003 Honda Accord",
		CarfaxRecords: 3,
	}
	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create record: %v", err)
	}

	transaction, err := create(db, user.ID, &record.ID, 2, 3400)
	if err != nil {
		t.Fatalf("create() error = %v", err)
	}
	if len(transaction.Token) != 36 {
		t.Fatalf("token length = %d, want 36", len(transaction.Token))
	}
	if _, err := uuid.Parse(transaction.Token); err != nil {
		t.Fatalf("token is not a UUID: %v", err)
	}
	if transaction.Status != domaine.TransactionPending {
		t.Fatalf("status = %q", transaction.Status)
	}
	if transaction.PaymentMethod != domaine.PaymentMethodPayPalCheckout {
		t.Fatalf("payment method = %q", transaction.PaymentMethod)
	}
	if transaction.Credits != 2 || transaction.Amount != 3400 {
		t.Fatalf("credits/amount = %d/%d", transaction.Credits, transaction.Amount)
	}

	found, err := findByToken(db, transaction.Token)
	if err != nil {
		t.Fatalf("findByToken() error = %v", err)
	}
	if found == nil || found.ID != transaction.ID {
		t.Fatal("findByToken() did not find the transaction")
	}
	if found.User.Email != user.Email {
		t.Fatalf("preloaded user email = %q", found.User.Email)
	}
	if found.CacheRecord == nil || found.CacheRecord.ID != record.ID {
		t.Fatal("preloaded cache record missing")
	}

	if found, _ := findByToken(db, "not-a-uuid"); found != nil {
		t.Fatal("malformed token matched a transaction")
	}
	if found, _ := findByToken(db, uuid.NewString()); found != nil {
		t.Fatal("missing UUID token matched a transaction")
	}
}
