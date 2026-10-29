package kernel

import "carf-chorf.com/app/pkg/config_loader"

var (
	Server         serverConfig
	Database       databaseConfig
	ReportProvider reportProviderConfig
	Discord        discordConfig
	Mail           mailConfig
	Turnstile      turnstileConfig
	Prices         pricesConfig
	PayPal         paypalConfig
	ReportStorage  reportStorageConfig
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
	config_loader.Register("DiscordConfig", &Discord)
	config_loader.Register("MailConfig", &Mail)
	config_loader.Register("TurnstileConfig", &Turnstile)
	config_loader.RegisterYAML("PricesConfig", "config/prices.yml", &Prices)
	config_loader.Register("PayPalConfig", &PayPal)
	config_loader.Register("ReportStorageConfig", &ReportStorage)
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

	// reportProviderConfig holds the credentials sent to carbpi.fr.
	reportProviderConfig struct {
		CarbpiTokenName   string `env:"CARBPI_TOKEN_NAME"`
		CarbpiTokenSecret string `env:"CARBPI_TOKEN_SECRET"`
	}

	discordConfig struct {
		ContactWebhookURL string `env:"DISCORD_CONTACT_WEBHOOK_URL"`
		SOSWebhookURL     string `env:"DISCORD_SOS_WEBHOOK_URL"`
	}

	mailConfig struct {
		Host     string `env:"MAIL_HOST"`
		Port     string `env:"MAIL_PORT"`
		Username string `env:"MAIL_USERNAME"`
		Password string `env:"MAIL_PASSWORD"`
		From     string `env:"MAIL_FROM"`
		FromName string `env:"MAIL_FROM_NAME"`
	}

	// turnstileConfig holds the Cloudflare Turnstile keys. The site key is
	// embedded in the lookup form; the secret key is only used server-side
	// when verifying the challenge response.
	turnstileConfig struct {
		SiteKey   string `env:"TURNSTILE_SITE_KEY"`
		SecretKey string `env:"TURNSTILE_SECRET_KEY"`
	}

	pricesConfig struct {
		Packages []PricePackage `yaml:"packages"`
	}

	// PricePackage is one purchasable credit bundle from config/prices.yml.
	PricePackage struct {
		Credits     uint   `yaml:"credits"`
		AmountCents uint   `yaml:"amount_cents"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Badge       string `yaml:"badge"`
	}

	paypalConfig struct {
		ClientID     string `env:"PAYPAL_CLIENT_ID"`
		ClientSecret string `env:"PAYPAL_CLIENT_SECRET"`
		APIURL       string `env:"PAYPAL_API_URL"`
	}

	reportStorageConfig struct {
		Path string `env:"REPORTS_PATH"`
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
