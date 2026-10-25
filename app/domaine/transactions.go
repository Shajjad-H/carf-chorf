package domaine

// TransactionStatus is the lifecycle state of a credit purchase.
type TransactionStatus string

const (
	TransactionPending TransactionStatus = "pending"
	TransactionPaid    TransactionStatus = "paid"
)

// PaymentMethod identifies the provider used to pay for a transaction.
type PaymentMethod string

const (
	PaymentMethodPayPalCheckout PaymentMethod = "paypal_checkout"
)

// Transaction records a credit package purchase. A transaction starts as
// pending and is marked paid after the payment provider confirms capture.
type Transaction struct {
	GormModel

	// Token is the public handle for the payment page.
	Token string `gorm:"column:token;type:varchar(36);uniqueIndex;not null" json:"token"`

	// Pid is the provider order/payment id, filled once the payment is created.
	Pid string `gorm:"column:pid;type:varchar(255);index" json:"pid"`

	Credits uint `gorm:"column:credits;not null" json:"credits"`
	Amount  uint `gorm:"column:amount;not null" json:"amount"` // cents, USD

	Status        TransactionStatus `gorm:"column:status;type:varchar(20);not null;default:'pending'" json:"status"`
	PaymentMethod PaymentMethod     `gorm:"column:payment_method;type:varchar(32);not null;default:'paypal_checkout'" json:"payment_method"`

	UserID uint `gorm:"column:user_id;not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user"`

	CacheRecordID *uint        `gorm:"column:cache_record_id;index" json:"cache_record_id,omitempty"`
	CacheRecord   *CacheRecord `gorm:"foreignKey:CacheRecordID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"cache_record,omitempty"`
}
