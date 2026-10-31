// Package admin_service provides dashboard queries and aggregates.
package admin_service

import (
	"strings"
	"time"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/database"
	"gorm.io/gorm"
)

// Summary is the dashboard overview.
type Summary struct {
	TotalTransactions int64
	PaidTransactions  int64
	TotalPaidAmount   uint64
	TotalUsers        int64
	UnusedCredits     uint64
	TotalOrders       int64
	TotalReports      int64
}

// TransactionFilters are optional filters for the transactions page.
type TransactionFilters struct {
	Email  string
	Token  string
	Pid    string
	Status string
	From   time.Time
	To     time.Time
}

// OrderFilters are optional filters for the orders page.
type OrderFilters struct {
	Email string
	VIN   string
}

// UserFilters are optional filters for the users page.
type UserFilters struct {
	Email      string
	Token      string
	MinCredits *uint
	Active     string
	From       time.Time
	To         time.Time
}

// DailyGroup is one amount group in the daily paid report.
type DailyGroup struct {
	Amount uint
	Count  int64
	Total  uint64
}

// GetSummary returns all dashboard headline metrics.
func GetSummary() (*Summary, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	var summary Summary
	if err := db.Model(&domaine.Transaction{}).Count(&summary.TotalTransactions).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&domaine.Transaction{}).
		Where("status = ?", domaine.TransactionPaid).
		Count(&summary.PaidTransactions).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&domaine.Transaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status = ?", domaine.TransactionPaid).
		Scan(&summary.TotalPaidAmount).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&domaine.User{}).Count(&summary.TotalUsers).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&domaine.User{}).
		Select("COALESCE(SUM(credits), 0)").
		Scan(&summary.UnusedCredits).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&domaine.Order{}).Count(&summary.TotalOrders).Error; err != nil {
		return nil, err
	}
	summary.TotalReports = summary.TotalOrders
	return &summary, nil
}

// ListTransactions returns a filtered page of transactions, the filtered count,
// and the paid amount for the same filter.
func ListTransactions(filters TransactionFilters, page, size int) ([]domaine.Transaction, int64, uint64, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, 0, 0, err
	}

	var total int64
	if err := transactionQuery(db, filters).Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	var paidAmount uint64
	if err := transactionQuery(db, filters).
		Where("transactions.status = ?", domaine.TransactionPaid).
		Select("COALESCE(SUM(transactions.amount), 0)").
		Scan(&paidAmount).Error; err != nil {
		return nil, 0, 0, err
	}

	var transactions []domaine.Transaction
	err = transactionQuery(db, filters).
		Preload("User").
		Order("transactions.created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, 0, err
	}
	return transactions, total, paidAmount, nil
}

func transactionQuery(db *gorm.DB, filters TransactionFilters) *gorm.DB {
	q := db.Model(&domaine.Transaction{}).
		Joins("JOIN users ON users.id = transactions.user_id AND users.deleted_at IS NULL")

	if filters.Email != "" {
		q = q.Where("LOWER(users.email) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(filters.Email))+"%")
	}
	if filters.Token != "" {
		q = q.Where("transactions.token = ?", strings.TrimSpace(filters.Token))
	}
	if filters.Pid != "" {
		q = q.Where("transactions.pid = ?", strings.TrimSpace(filters.Pid))
	}
	if filters.Status != "" {
		q = q.Where("transactions.status = ?", filters.Status)
	}
	if !filters.From.IsZero() {
		q = q.Where("transactions.created_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		q = q.Where("transactions.created_at < ?", filters.To.AddDate(0, 0, 1))
	}
	return q
}

// ListOrders returns a filtered page of report orders.
func ListOrders(filters OrderFilters, page, size int) ([]domaine.Order, int64, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, 0, err
	}

	q := orderQuery(db, filters)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []domaine.Order
	err = orderQuery(db, filters).
		Preload("User").
		Preload("CacheRecord").
		Order("orders.created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func orderQuery(db *gorm.DB, filters OrderFilters) *gorm.DB {
	q := db.Model(&domaine.Order{}).
		Joins("JOIN users ON users.id = orders.user_id AND users.deleted_at IS NULL")

	if filters.Email != "" {
		q = q.Where("LOWER(users.email) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(filters.Email))+"%")
	}
	if filters.VIN != "" {
		q = q.Where("LOWER(orders.vin) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(filters.VIN))+"%")
	}
	return q
}

// DailyReport returns today's paid transactions grouped by amount.
func DailyReport(day time.Time) ([]DailyGroup, uint64, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, 0, err
	}

	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.AddDate(0, 0, 1)

	var groups []DailyGroup
	err = db.Model(&domaine.Transaction{}).
		Select("amount, COUNT(*) AS count, COALESCE(SUM(amount), 0) AS total").
		Where("status = ? AND created_at >= ? AND created_at < ?", domaine.TransactionPaid, start, end).
		Group("amount").
		Order("amount").
		Scan(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	var total uint64
	if err := db.Model(&domaine.Transaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status = ? AND created_at >= ? AND created_at < ?", domaine.TransactionPaid, start, end).
		Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

// ListUsers returns a filtered page of users.
func ListUsers(filters UserFilters, page, size int) ([]domaine.User, int64, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := userQuery(db, filters).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []domaine.User
	err = userQuery(db, filters).
		Order("users.created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func userQuery(db *gorm.DB, filters UserFilters) *gorm.DB {
	q := db.Model(&domaine.User{})

	if filters.Email != "" {
		q = q.Where("LOWER(users.email) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(filters.Email))+"%")
	}
	if filters.Token != "" {
		q = q.Where("EXISTS (SELECT 1 FROM transactions WHERE transactions.user_id = users.id AND transactions.token = ? AND transactions.deleted_at IS NULL)", strings.TrimSpace(filters.Token))
	}
	if filters.MinCredits != nil {
		q = q.Where("users.credits >= ?", *filters.MinCredits)
	}
	switch filters.Active {
	case "active":
		q = q.Where("users.is_active = ?", true)
	case "inactive":
		q = q.Where("users.is_active = ?", false)
	}
	if !filters.From.IsZero() {
		q = q.Where("users.created_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		q = q.Where("users.created_at < ?", filters.To.AddDate(0, 0, 1))
	}
	return q
}

// CountReminderRecipients counts active users with unused credits matching filters.
func CountReminderRecipients(filters UserFilters) (int64, error) {
	db, err := database.GetDB()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := reminderQuery(db, filters).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ListReminderRecipients returns all active users with unused credits matching filters.
func ListReminderRecipients(filters UserFilters) ([]domaine.User, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	var users []domaine.User
	err = reminderQuery(db, filters).Order("users.created_at DESC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func reminderQuery(db *gorm.DB, filters UserFilters) *gorm.DB {
	return userQuery(db, filters).Where("users.is_active = ? AND users.credits > 0", true)
}
