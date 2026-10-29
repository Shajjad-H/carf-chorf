package kernel

import (
	"errors"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

func (c *databaseConfig) Validate() error {
	if err := validation.ValidateStruct(c,
		validation.Field(&c.Driver,
			validation.Required.Error("DB_DRIVER is required"),
			validation.In(DatabaseMySQL, DatabasePostgreSQL, DatabaseSQLite).
				Error("DB_DRIVER must be one of: mysql, postgres, sqlite"),
		),
	); err != nil {
		return err
	}
	switch c.Driver {
	case DatabaseSQLite:
		// SQLite only requires DB_NAME (file path)
		return validation.ValidateStruct(c,
			validation.Field(&c.Name, validation.Required.Error("DB_NAME is required for SQLite")),
		)
	case DatabaseMySQL:
		// MySQL requires DB_HOST, DB_PORT, DB_NAME, DB_USER, and DB_PASSWORD
		return validation.ValidateStruct(c,
			validation.Field(&c.Host, validation.Required.Error("DB_HOST is required for MySQL")),
			validation.Field(&c.Port, validation.Required.Error("DB_PORT is required for MySQL")),
			validation.Field(&c.Name, validation.Required.Error("DB_NAME is required for MySQL")),
			validation.Field(&c.User, validation.Required.Error("DB_USER is required for MySQL")),
			validation.Field(&c.Password, validation.Required.Error("DB_PASSWORD is required for MySQL")),
		)
	case DatabasePostgreSQL:
		// PostgreSQL requires DB_HOST, DB_PORT, DB_NAME, DB_USER, and DB_PASSWORD
		return validation.ValidateStruct(c,
			validation.Field(&c.Host, validation.Required.Error("DB_HOST is required for PostgreSQL")),
			validation.Field(&c.Port, validation.Required.Error("DB_PORT is required for PostgreSQL")),
			validation.Field(&c.Name, validation.Required.Error("DB_NAME is required for PostgreSQL")),
			validation.Field(&c.User, validation.Required.Error("DB_USER is required for PostgreSQL")),
			validation.Field(&c.Password, validation.Required.Error("DB_PASSWORD is required for PostgreSQL")),
		)
	}
	return nil
}

func (v *serverConfig) Validate() error {
	return validation.ValidateStruct(v,
		validation.Field(&v.Address, validation.Required.Error("SERVER_ADDRESS is required")),
		validation.Field(&v.Env,
			validation.Required.Error("APP_ENV is required"),
			validation.In("dev", "prod").Error("APP_ENV must be one of: dev, prod"),
		),
		validation.Field(&v.Domain,
			validation.Required.Error("APP_DOMAIN is required"),
			is.URL.Error("APP_DOMAIN must be a valid URL"),
		),
		validation.Field(&v.SessionSecret,
			validation.Required.Error("SESSION_SECRET is required"),
			validation.Length(32, 0).Error("SESSION_SECRET must be at least 32 characters"),
		),
	)
}

func (v *reportProviderConfig) Validate() error {
	return validation.ValidateStruct(v,
		validation.Field(&v.CarbpiTokenName, validation.Required.Error("CARBPI_TOKEN_NAME is required")),
		validation.Field(&v.CarbpiTokenSecret, validation.Required.Error("CARBPI_TOKEN_SECRET is required")),
	)
}

func (d *discordConfig) Validate() error {
	return validation.ValidateStruct(d,
		validation.Field(&d.ContactWebhookURL,
			validation.Required.Error("DISCORD_CONTACT_WEBHOOK_URL is required"),
			is.URL.Error("DISCORD_CONTACT_WEBHOOK_URL must be a valid URL"),
		),
		validation.Field(&d.SOSWebhookURL,
			validation.Required.Error("DISCORD_SOS_WEBHOOK_URL is required"),
			is.URL.Error("DISCORD_SOS_WEBHOOK_URL must be a valid URL"),
		),
	)
}

func (t *turnstileConfig) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.SiteKey, validation.Required.Error("TURNSTILE_SITE_KEY is required")),
		validation.Field(&t.SecretKey, validation.Required.Error("TURNSTILE_SECRET_KEY is required")),
	)
}

func (m *mailConfig) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Host, validation.Required.Error("MAIL_HOST is required")),
		validation.Field(&m.Port,
			validation.Required.Error("MAIL_PORT is required"),
			is.Port.Error("MAIL_PORT must be a valid port number"),
		),
		validation.Field(&m.Username, validation.Required.Error("MAIL_USERNAME is required")),
		validation.Field(&m.Password, validation.Required.Error("MAIL_PASSWORD is required")),
		validation.Field(&m.From,
			validation.Required.Error("MAIL_FROM is required"),
			is.Email.Error("MAIL_FROM must be a valid email address"),
		),
	)
}

func (p *paypalConfig) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.ClientID, validation.Required.Error("PAYPAL_CLIENT_ID is required")),
		validation.Field(&p.ClientSecret, validation.Required.Error("PAYPAL_CLIENT_SECRET is required")),
		validation.Field(&p.APIURL,
			validation.Required.Error("PAYPAL_API_URL is required"),
			is.URL.Error("PAYPAL_API_URL must be a valid URL"),
		),
	)
}

func (r *reportStorageConfig) Validate() error {
	return nil
}

func (p *pricesConfig) Validate() error {
	if len(p.Packages) == 0 {
		return errors.New("at least one package is required")
	}

	seen := make(map[uint]struct{}, len(p.Packages))
	hasSingleCredit := false

	for i, pkg := range p.Packages {
		if pkg.Credits == 0 {
			return fmt.Errorf("packages[%d].credits must be greater than zero", i)
		}
		if pkg.AmountCents == 0 {
			return fmt.Errorf("packages[%d].amount_cents must be greater than zero", i)
		}
		if strings.TrimSpace(pkg.Name) == "" {
			return fmt.Errorf("packages[%d].name is required", i)
		}
		if _, exists := seen[pkg.Credits]; exists {
			return fmt.Errorf("packages[%d].credits %d is duplicated", i, pkg.Credits)
		}

		seen[pkg.Credits] = struct{}{}
		if pkg.Credits == 1 {
			hasSingleCredit = true
		}
	}

	if !hasSingleCredit {
		return errors.New("a 1-credit package is required")
	}
	return nil
}
