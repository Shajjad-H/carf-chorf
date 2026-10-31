package admin_service

import (
	"path/filepath"
	"testing"
	"time"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/app/services/database"
	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	kernel.Database.Driver = kernel.DatabaseSQLite
	kernel.Database.Name = filepath.Join(t.TempDir(), "admin.db")

	db, err := database.GetDB()
	if err != nil {
		t.Fatalf("get database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.User{}, &domaine.CacheRecord{}, &domaine.Transaction{}, &domaine.Order{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func TestAdminQueries(t *testing.T) {
	setupTestDB(t)
	db, _ := database.GetDB()

	user1 := &domaine.User{Email: "one@example.com", Credits: 3, IsActive: true}
	user2 := &domaine.User{Email: "two@example.com", Credits: 0, IsActive: true}
	user3 := &domaine.User{Email: "three@example.com", Credits: 5, IsActive: false}
	for _, user := range []*domaine.User{user1, user2, user3} {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	if err := db.Model(user3).Update("is_active", false).Error; err != nil {
		t.Fatalf("deactivate user: %v", err)
	}

	record := &domaine.CacheRecord{
		Token:    uuid.NewString(),
		Vin:      "1HGCM82633A004352",
		Provider: domaine.CarbpiProvider,
		Vehicle:  "2003 Honda Accord",
	}
	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create record: %v", err)
	}

	now := time.Now()
	paid1 := &domaine.Transaction{
		Token: uuid.NewString(), Pid: "pid-1", Credits: 1, Amount: 1900,
		Status: domaine.TransactionPaid, PaymentMethod: domaine.PaymentMethodPayPalCheckout,
		UserID: user1.ID, GormModel: domaine.GormModel{CreatedAt: now},
	}
	pending := &domaine.Transaction{
		Token: uuid.NewString(), Pid: "pid-2", Credits: 2, Amount: 3400,
		Status: domaine.TransactionPending, PaymentMethod: domaine.PaymentMethodPayPalCheckout,
		UserID: user2.ID, GormModel: domaine.GormModel{CreatedAt: now},
	}
	paid2 := &domaine.Transaction{
		Token: uuid.NewString(), Pid: "pid-3", Credits: 2, Amount: 3400,
		Status: domaine.TransactionPaid, PaymentMethod: domaine.PaymentMethodPayPalCheckout,
		UserID: user3.ID, GormModel: domaine.GormModel{CreatedAt: now},
	}
	for _, transaction := range []*domaine.Transaction{paid1, pending, paid2} {
		if err := db.Create(transaction).Error; err != nil {
			t.Fatalf("create transaction: %v", err)
		}
	}

	for _, order := range []*domaine.Order{
		{UserID: user1.ID, CacheRecordID: record.ID, Vin: record.Vin, AccessToken: uuid.NewString()},
		{UserID: user3.ID, CacheRecordID: record.ID, Vin: record.Vin, AccessToken: uuid.NewString()},
	} {
		if err := db.Create(order).Error; err != nil {
			t.Fatalf("create order: %v", err)
		}
	}

	summary, err := GetSummary()
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if summary.TotalTransactions != 3 || summary.PaidTransactions != 2 {
		t.Fatalf("transaction summary = %+v", summary)
	}
	if summary.TotalPaidAmount != 5300 {
		t.Fatalf("paid amount = %d, want 5300", summary.TotalPaidAmount)
	}
	if summary.TotalUsers != 3 || summary.UnusedCredits != 8 {
		t.Fatalf("user summary = %+v", summary)
	}
	if summary.TotalOrders != 2 || summary.TotalReports != 2 {
		t.Fatalf("order summary = %+v", summary)
	}

	transactions, total, paidAmount, err := ListTransactions(TransactionFilters{Status: "paid"}, 1, 25)
	if err != nil {
		t.Fatalf("ListTransactions() error = %v", err)
	}
	if total != 2 || paidAmount != 5300 || len(transactions) != 2 {
		t.Fatalf("paid transactions = %d/%d/%d", total, paidAmount, len(transactions))
	}

	transactions, total, _, err = ListTransactions(TransactionFilters{Email: "one@"}, 1, 25)
	if err != nil {
		t.Fatalf("ListTransactions(email) error = %v", err)
	}
	if total != 1 || len(transactions) != 1 || transactions[0].ID != paid1.ID {
		t.Fatal("email filter did not return the expected transaction")
	}

	transactions, total, _, err = ListTransactions(TransactionFilters{Token: pending.Token}, 1, 25)
	if err != nil {
		t.Fatalf("ListTransactions(token) error = %v", err)
	}
	if total != 1 || transactions[0].ID != pending.ID {
		t.Fatal("token filter did not return the expected transaction")
	}

	orders, total, err := ListOrders(OrderFilters{Email: "one@"}, 1, 25)
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	if total != 1 || len(orders) != 1 {
		t.Fatalf("orders = %d/%d", total, len(orders))
	}

	groups, dailyTotal, err := DailyReport(now)
	if err != nil {
		t.Fatalf("DailyReport() error = %v", err)
	}
	if dailyTotal != 5300 || len(groups) != 2 {
		t.Fatalf("daily report = %d/%d", dailyTotal, len(groups))
	}

	users, total, err := ListUsers(UserFilters{Active: "active"}, 1, 25)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Fatalf("active users = %d/%d", total, len(users))
	}

	users, total, err = ListUsers(UserFilters{Token: pending.Token}, 1, 25)
	if err != nil {
		t.Fatalf("ListUsers(token) error = %v", err)
	}
	if total != 1 || users[0].ID != user2.ID {
		t.Fatal("token filter did not return the expected user")
	}

	recipients, err := ListReminderRecipients(UserFilters{})
	if err != nil {
		t.Fatalf("ListReminderRecipients() error = %v", err)
	}
	if len(recipients) != 1 || recipients[0].ID != user1.ID {
		t.Fatalf("recipients = %+v", recipients)
	}

	count, err := CountReminderRecipients(UserFilters{})
	if err != nil {
		t.Fatalf("CountReminderRecipients() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("reminder count = %d, want 1", count)
	}

	count, err = CountReminderRecipients(UserFilters{Active: "inactive"})
	if err != nil {
		t.Fatalf("CountReminderRecipients(inactive) error = %v", err)
	}
	if count != 0 {
		t.Fatalf("inactive reminder count = %d, want 0", count)
	}

	count, err = CountReminderRecipients(UserFilters{Email: "one@"})
	if err != nil {
		t.Fatalf("CountReminderRecipients(email) error = %v", err)
	}
	if count != 1 {
		t.Fatalf("email reminder count = %d, want 1", count)
	}
}
