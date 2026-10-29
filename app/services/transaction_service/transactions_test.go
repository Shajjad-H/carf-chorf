package transaction_service

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

func openCompleteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := filepath.Join(t.TempDir(), "complete.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.User{}, &domaine.CacheRecord{}, &domaine.Transaction{}, &domaine.Order{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCompletePurchaseWithVehicle(t *testing.T) {
	db := openCompleteTestDB(t)

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
		t.Fatalf("create transaction: %v", err)
	}

	reportToken := uuid.NewString()
	completed, order, orderCreated, err := completePurchase(db, transaction.Token, "paypal-1", reportToken)
	if err != nil {
		t.Fatalf("completePurchase() error = %v", err)
	}
	if !orderCreated || order == nil || order.AccessToken == "" {
		t.Fatal("expected a new report order")
	}
	if completed.Status != domaine.TransactionPaid {
		t.Fatalf("status = %q", completed.Status)
	}
	if completed.Pid != "paypal-1" {
		t.Fatalf("pid = %q", completed.Pid)
	}
	if order.TransactionID == nil || *order.TransactionID != transaction.ID || order.Vin != record.Vin || order.CacheRecordID != record.ID {
		t.Fatalf("unexpected order: %+v", order)
	}
	if order.AccessToken != reportToken {
		t.Fatalf("access token = %q, want %q", order.AccessToken, reportToken)
	}

	var updatedUser domaine.User
	if err := db.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if updatedUser.Credits != 1 {
		t.Fatalf("credits = %d, want 1", updatedUser.Credits)
	}

	completed, secondOrder, orderCreated, err := completePurchase(db, transaction.Token, "paypal-1", "")
	if err != nil {
		t.Fatalf("second completePurchase() error = %v", err)
	}
	if orderCreated {
		t.Fatal("second call created another order")
	}
	if secondOrder.ID != order.ID {
		t.Fatal("second call returned a different order")
	}
	if completed.Status != domaine.TransactionPaid {
		t.Fatalf("status = %q", completed.Status)
	}

	if err := db.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if updatedUser.Credits != 1 {
		t.Fatalf("credits after retry = %d, want 1", updatedUser.Credits)
	}

	var orderCount int64
	if err := db.Model(&domaine.Order{}).Count(&orderCount).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if orderCount != 1 {
		t.Fatalf("order count = %d, want 1", orderCount)
	}
}

func TestCompletePurchaseWithoutVehicle(t *testing.T) {
	db := openCompleteTestDB(t)

	user := &domaine.User{Email: "user@example.com", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	transaction, err := create(db, user.ID, nil, 2, 3400)
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	completed, order, orderCreated, err := completePurchase(db, transaction.Token, "paypal-2", "")
	if err != nil {
		t.Fatalf("completePurchase() error = %v", err)
	}
	if completed.Status != domaine.TransactionPaid {
		t.Fatalf("status = %q", completed.Status)
	}
	if orderCreated || order.AccessToken != "" {
		t.Fatal("unexpected report order")
	}

	var updatedUser domaine.User
	if err := db.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if updatedUser.Credits != 2 {
		t.Fatalf("credits = %d, want 2", updatedUser.Credits)
	}
}

func TestCompletePurchaseRejectsPidMismatch(t *testing.T) {
	db := openCompleteTestDB(t)

	user := &domaine.User{Email: "user@example.com", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	transaction, err := create(db, user.ID, nil, 1, 1900)
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if err := db.Model(transaction).Update("pid", "paypal-a").Error; err != nil {
		t.Fatalf("set pid: %v", err)
	}

	_, _, _, err = completePurchase(db, transaction.Token, "paypal-b", "")
	if !errors.Is(err, ErrTransactionPidMismatch) {
		t.Fatalf("error = %v, want ErrTransactionPidMismatch", err)
	}
}

func TestCompletePurchaseGeneratesReportToken(t *testing.T) {
	db := openCompleteTestDB(t)

	user := &domaine.User{Email: "generated@example.com", IsActive: true}
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
	transaction, err := create(db, user.ID, &record.ID, 1, 1900)
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	_, order, orderCreated, err := completePurchase(db, transaction.Token, "paypal-token", "")
	if err != nil {
		t.Fatalf("completePurchase() error = %v", err)
	}
	if !orderCreated || order == nil {
		t.Fatal("expected a report order")
	}
	if _, err := uuid.Parse(order.AccessToken); err != nil {
		t.Fatalf("generated access token is not a UUID: %v", err)
	}
}

func TestOrderAllowsNilTransactionID(t *testing.T) {
	db := openCompleteTestDB(t)

	user := &domaine.User{Email: "credits@example.com", IsActive: true}
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

	first := &domaine.Order{
		UserID:        user.ID,
		CacheRecordID: record.ID,
		Vin:           record.Vin,
		AccessToken:   uuid.NewString(),
	}
	second := &domaine.Order{
		UserID:        user.ID,
		CacheRecordID: record.ID,
		Vin:           record.Vin,
		AccessToken:   uuid.NewString(),
	}

	if err := db.Create(first).Error; err != nil {
		t.Fatalf("create first order: %v", err)
	}
	if err := db.Create(second).Error; err != nil {
		t.Fatalf("create second order: %v", err)
	}

	var count int64
	if err := db.Model(&domaine.Order{}).Count(&count).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if count != 2 {
		t.Fatalf("order count = %d, want 2", count)
	}
}

func TestListByUser(t *testing.T) {
	db := openCompleteTestDB(t)

	user := &domaine.User{Email: "list@example.com", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := create(db, user.ID, nil, 1, 1900); err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if _, err := create(db, user.ID, nil, 2, 3400); err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	transactions, err := listByUser(db, user.ID)
	if err != nil {
		t.Fatalf("listByUser() error = %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("transactions = %d, want 2", len(transactions))
	}
}

func TestListPaidByUser(t *testing.T) {
	db := openCompleteTestDB(t)

	user := &domaine.User{Email: "paid@example.com", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	paid, err := create(db, user.ID, nil, 1, 1900)
	if err != nil {
		t.Fatalf("create paid transaction: %v", err)
	}
	if err := db.Model(paid).Update("status", domaine.TransactionPaid).Error; err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if _, err := create(db, user.ID, nil, 2, 3400); err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	transactions, err := listPaidByUser(db, user.ID)
	if err != nil {
		t.Fatalf("listPaidByUser() error = %v", err)
	}
	if len(transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(transactions))
	}
	if transactions[0].ID != paid.ID {
		t.Fatalf("transaction ID = %d, want %d", transactions[0].ID, paid.ID)
	}
}
