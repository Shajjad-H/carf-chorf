package kernel

import "carf-chorf.com/app/pkg/config_loader"

var (
	Server         serverConfig
	Database       databaseConfig
	ReportProvider reportProviderConfig
)

type DatabaseDriver string

const (
	DatabaseMySQL      DatabaseDriver = "mysql"
	DatabasePostgreSQL DatabaseDriver = "postgres"
	DatabaseSQLite     DatabaseDriver = "sqlite"
)

func Init() {

	config_loader.Register("ServerConfig", &Server)
	config_loader.Register("DatabaseConfig", &Database)
	config_loader.Register("ReportProvider", &ReportProvider)
	// config_loader.Register("StorageConfig", &Storage)
	// config_loader.Register("DashboardConfig", &Dashboard)

	config_loader.LoadAndValidate(true)
}

func IsDev() bool {
	return Server.Env == "dev"
}

type (
	serverConfig struct {
		Address string `env:"SERVER_ADDRESS"`
		Env     string `env:"APP_ENV"`
		// Domain is the public base URL (e.g. "https://example.com"), used to
		// build absolute asset URLs for contexts that can't resolve relative
		// paths — namely the logo in transactional emails, since mail clients
		// load images straight from the public internet, not from this server.
		Domain        string `env:"APP_DOMAIN"`
		SessionSecret string `env:"SESSION_SECRET"`
	}

	databaseConfig struct {
		Driver   DatabaseDriver `env:"DB_DRIVER"`
		Host     string         `env:"DB_HOST"`
		Port     string         `env:"DB_PORT"`
		Name     string         `env:"DB_NAME"`
		User     string         `env:"DB_USER"`
		Password string         `env:"DB_PASSWORD"`
		Migrate  bool           `env:"MIGRATE_DB"`
	}

	reportProviderConfig struct {
		CarbpiTokenName   string `env:"CARBPI_TOKEN_NAME"`
		CarbpiTokenSecret string `env:"CARBPI_TOKEN_SECRET"`
	}
)

// type (
// 	dashboardConfig struct {
// 		DASHBOARD_ADDRESS  string `validate:"required" env:"DASHBOARD_ADDRESS"` // e.g. :8081
// 		DASHBOARD_USER     string `validate:"required" env:"DASHBOARD_USER"`
// 		DASHBOARD_PASSWORD string `validate:"required" env:"DASHBOARD_PASSWORD"`
// 		SESSION_SECRET     string `validate:"required" env:"SESSION_SECRET"`
// 		// comma-separated, e.g. localhost:8081,dash.carbpi.fr
// 		DASHBOARD_ORIGIN string `env:"DASHBOARD_ORIGIN"`
// 	}
// )
