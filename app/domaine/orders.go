package domaine

// Order is a purchased vehicle report. AccessToken is the public handle used
// by the report viewer. TransactionID links payment-created orders to their
// transaction and is nil for reports bought later with credits.
type Order struct {
	GormModel

	UserID uint `gorm:"column:user_id;not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user"`

	CacheRecordID uint        `gorm:"column:cache_record_id;not null;index" json:"cache_record_id"`
	CacheRecord   CacheRecord `gorm:"foreignKey:CacheRecordID" json:"cache_record"`

	TransactionID *uint  `gorm:"column:transaction_id;index" json:"transaction_id,omitempty"`
	Vin           string `gorm:"column:vin;type:varchar(20);not null;index" json:"vin"`
	AccessToken   string `gorm:"column:access_token;type:varchar(36);uniqueIndex;not null" json:"access_token"`
}
