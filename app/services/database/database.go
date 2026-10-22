package database

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/simple_logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
	dbErr  error
)

func GetDB() (*gorm.DB, error) {
	dbOnce.Do(func() { db, dbErr = openDB() })
	return db, dbErr
}

// func Migrate() error {
// 	g, err := GetDB()
// 	if err != nil {
// 		return err
// 	}
// 	for _, m := range domaine.MigrationModels() {
// 		simple_logger.Infof("migrating: %v", m)
// 		if err := g.AutoMigrate(m); err != nil {
// 			return err
// 		}
// 	}
// 	// legacy cleanup: vin_provider moved from Vehicle to Registration
// 	if g.Migrator().HasColumn(&domaine.Vehicle{}, "vin_provider") {
// 		if err := g.Migrator().DropColumn(&domaine.Vehicle{}, "vin_provider"); err != nil {
// 			return err
// 		}
// 		simple_logger.Infof("dropped legacy column vehicles.vin_provider")
// 	}

// 	return nil
// }

func openDB() (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch kernel.DatabaseDriver(kernel.Database.Driver) {
	case kernel.DatabasePostgreSQL:
		dialector = postgres.Open(postgresDSN())
	case kernel.DatabaseMySQL:
		dialector = mysql.Open(mysqlDSN())
	case kernel.DatabaseSQLite:
		if err := os.MkdirAll(filepath.Dir(kernel.Database.Name), 0o755); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(kernel.Database.Name)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", kernel.Database.Driver)
	}

	logLevel := logger.Warn
	if kernel.IsDev() {
		logLevel = logger.Info
	}

	g, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := g.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	simple_logger.Successf("database connected [%s]", kernel.Database.Driver)
	return g, nil
}

func postgresDSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(kernel.Database.User, kernel.Database.Password),
		Host:   net.JoinHostPort(kernel.Database.Host, kernel.Database.Port),
		Path:   "/" + kernel.Database.Name,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func mysqlDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		kernel.Database.User, kernel.Database.Password, kernel.Database.Host, kernel.Database.Port, kernel.Database.Name)
}
