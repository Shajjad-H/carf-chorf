package vehicle

import (
	"path/filepath"
	"testing"

	"carf-chorf.com/app/app/domaine"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewToken(t *testing.T) {
	first, err := newToken()
	if err != nil {
		t.Fatalf("newToken() error = %v", err)
	}
	if len(first) != 36 {
		t.Fatalf("token length = %d, want 36", len(first))
	}
	id, err := uuid.Parse(first)
	if err != nil {
		t.Fatalf("token is not a UUID: %v", err)
	}
	if id.Version() != 4 {
		t.Fatalf("UUID version = %d, want 4", id.Version())
	}

	second, err := newToken()
	if err != nil {
		t.Fatalf("newToken() error = %v", err)
	}
	if first == second {
		t.Fatal("two tokens were identical")
	}
}

func TestFindRecordByInvalidToken(t *testing.T) {
	for _, token := range []string{
		"too-short",
		"00000000-0000-0000-0000-00000000000z",
	} {
		record, err := FindRecordByToken(token)
		if err != nil {
			t.Fatalf("FindRecordByToken(%q) error = %v", token, err)
		}
		if record != nil {
			t.Fatalf("expected nil record for invalid token %q", token)
		}
	}
}

func TestFirstOrNil(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "first.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.CacheRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	missing, err := firstOrNil(db.Where("vin = ?", "MISSING"))
	if err != nil {
		t.Fatalf("firstOrNil() error = %v", err)
	}
	if missing != nil {
		t.Fatal("expected nil for a missing row")
	}

	created := &domaine.CacheRecord{
		Token:         uuid.NewString(),
		Vin:           "1HGCM82633A004352",
		Provider:      domaine.CarbpiProvider,
		Vehicle:       "2003 Honda Accord",
		CarfaxRecords: 2,
	}
	if err := db.Create(created).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	found, err := firstOrNil(db.Where("vin = ?", created.Vin))
	if err != nil {
		t.Fatalf("firstOrNil() error = %v", err)
	}
	if found == nil || found.Token != created.Token {
		t.Fatal("firstOrNil() did not return the created record")
	}
}

func TestFillMissingPlateState(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "vehicle.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.CacheRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	existing := &domaine.CacheRecord{
		Token:         uuid.NewString(),
		Vin:           "1HGCM82633A004352",
		Provider:      domaine.CarbpiProvider,
		Vehicle:       "2003 Honda Accord",
		CarfaxRecords: 3,
	}
	if err := db.Create(existing).Error; err != nil {
		t.Fatalf("create existing: %v", err)
	}

	incoming := &domaine.CacheRecord{
		Vin:           existing.Vin,
		Provider:      existing.Provider,
		Plate:         "ABC123",
		State:         "CA",
		Vehicle:       "Different Vehicle",
		CarfaxRecords: 9,
	}

	updated, err := fillMissingPlateState(db, existing, incoming)
	if err != nil {
		t.Fatalf("fillMissingPlateState() error = %v", err)
	}
	if updated.Token != existing.Token {
		t.Fatal("token changed")
	}
	if updated.Plate != "ABC123" || updated.State != "CA" {
		t.Fatalf("plate/state = %q/%q", updated.Plate, updated.State)
	}
	if updated.Vehicle != "2003 Honda Accord" || updated.CarfaxRecords != 3 {
		t.Fatal("non-plate fields were overwritten")
	}

	secondIncoming := &domaine.CacheRecord{
		Vin:      existing.Vin,
		Provider: existing.Provider,
		Plate:    "ZZZ999",
		State:    "NY",
	}
	updated, err = fillMissingPlateState(db, updated, secondIncoming)
	if err != nil {
		t.Fatalf("fillMissingPlateState() error = %v", err)
	}
	if updated.Plate != "ABC123" || updated.State != "CA" {
		t.Fatal("existing plate/state was overwritten")
	}

	var count int64
	if err := db.Model(&domaine.CacheRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("record count = %d, want 1", count)
	}
}
