package domaine

import (
	"time"

	"gorm.io/gorm"
)

type GormModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type User struct {
	GormModel
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Credits  uint   `gorm:"not null;default:0" json:"credits"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}

// ReportProvider identifies the upstream service that produced a cached report.
type ReportProvider string

const (
	// CarbpiProvider is the carbpi.fr vehicle-history API.
	CarbpiProvider ReportProvider = "carbpi"
)

// CacheRecord is a cached provider response for one vehicle. Rows are unique
// per (vin, provider), and Token is the public handle used by the package page
// so the VIN does not appear in URLs.
type CacheRecord struct {
	GormModel

	// Token is a random UUIDv4 used to look up the record without exposing
	// the VIN in links.
	Token string `gorm:"column:token;type:varchar(36);uniqueIndex;not null" json:"-"`

	// Vin and Provider form the cache key; the same VIN may exist for
	// different providers.
	Vin      string         `gorm:"column:vin;type:varchar(20);not null;uniqueIndex:idx_vin_provider" json:"vin"`
	Provider ReportProvider `gorm:"column:provider;type:varchar(20);not null;default:'carbpi';uniqueIndex:idx_vin_provider" json:"provider"`

	// Plate and State are the last values known for the vehicle. They are
	// informational and may be empty when the lookup was VIN-only. The
	// composite index supports the plate/state cache lookup.
	Plate string `gorm:"column:plate;type:varchar(15);index:idx_plate_state" json:"plate"`
	State string `gorm:"column:state;type:varchar(50);index:idx_plate_state" json:"state"`

	// Vehicle is the provider's combined year/make/model display string.
	Vehicle string `gorm:"column:vehicle;type:varchar(255);not null;default:''" json:"vehicle"`

	// CarfaxRecords is the number of available CARFAX records. A provider
	// null is stored as zero because the lookup UI treats both as "none".
	CarfaxRecords uint `gorm:"column:carfax_records;not null;default:0" json:"carfax_records"`
}

// MigrationModels returns every model that should be auto-migrated.
func MigrationModels() []any {
	return []any{&User{}, &CacheRecord{}, &Transaction{}}
}
