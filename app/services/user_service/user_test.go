package user_service

import (
	"path/filepath"
	"testing"

	"carf-chorf.com/app/app/domaine"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUserLookupAndCreate(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "users.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domaine.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	created, err := create(db, "  User@Example.COM ")
	if err != nil {
		t.Fatalf("create() error = %v", err)
	}
	if created.Email != "user@example.com" || !created.IsActive {
		t.Fatalf("created user = %+v", created)
	}

	found, err := findByEmail(db, "USER@example.com")
	if err != nil {
		t.Fatalf("findByEmail() error = %v", err)
	}
	if found == nil || found.ID != created.ID {
		t.Fatal("findByEmail() did not find the created user")
	}

	byID, err := findByID(db, created.ID)
	if err != nil {
		t.Fatalf("findByID() error = %v", err)
	}
	if byID == nil || byID.Email != created.Email {
		t.Fatal("findByID() did not find the created user")
	}

	duplicate, err := create(db, "user@example.com")
	if err != nil {
		t.Fatalf("create() duplicate error = %v", err)
	}
	if duplicate.ID != created.ID {
		t.Fatal("duplicate create did not return the existing user")
	}

	other, err := findOrCreateByEmail(db, "other@example.com")
	if err != nil {
		t.Fatalf("findOrCreateByEmail() error = %v", err)
	}
	if other.ID == created.ID {
		t.Fatal("findOrCreateByEmail() returned the wrong user")
	}
}
