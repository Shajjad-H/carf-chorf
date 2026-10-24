package kernel

import (
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
