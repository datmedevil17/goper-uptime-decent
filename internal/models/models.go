package models

import (
	"time"
)

// User model
type User struct {
	ID       string `gorm:"primaryKey;type:varchar(255)"`
	Email    string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (User) TableName() string {
	return "User"
}

// Website model
type Website struct {
	ID        string        `gorm:"primaryKey;type:varchar(255)"`
	URL       string        `gorm:"type:varchar(500);not null"`
	UserID    string        `gorm:"type:varchar(255);not null;index"`
	Disabled  bool          `gorm:"default:false"`
	Ticks     []WebsiteTick `gorm:"foreignKey:WebsiteID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Website) TableName() string {
	return "Website"
}

// Validator model
type Validator struct {
	ID             string        `gorm:"primaryKey;type:varchar(255)"`
	PublicKey      string        `gorm:"type:varchar(255);not null;uniqueIndex"`
	Location       string        `gorm:"type:varchar(255)"`
	IP             string        `gorm:"type:varchar(255)"`
	PendingPayouts float64       `gorm:"type:decimal(20,2);default:0"`
	Ticks          []WebsiteTick `gorm:"foreignKey:ValidatorID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Validator) TableName() string {
	return "Validator"
}

// WebsiteTick model
type WebsiteTick struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	WebsiteID   string    `gorm:"type:varchar(255);not null;index"`
	ValidatorID string    `gorm:"type:varchar(255);not null;index"`
	Status      string    `gorm:"type:varchar(50);not null"` // Good or Bad
	Latency     float64   `gorm:"type:decimal(10,2)"`
	CreatedAt   time.Time `gorm:"index"`

	Website   *Website   `gorm:"foreignKey:WebsiteID;constraint:OnDelete:CASCADE" json:",omitempty"`
	Validator *Validator `gorm:"foreignKey:ValidatorID;constraint:OnDelete:CASCADE" json:",omitempty"`
}

func (WebsiteTick) TableName() string {
	return "WebsiteTick"
}

// PayoutTransaction model
type PayoutTransaction struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)"`
	ValidatorID  string    `gorm:"type:varchar(255);not null;index"`
	Amount       float64   `gorm:"type:decimal(20,2);not null"`
	Status       string    `gorm:"type:varchar(50);not null;index"` // pending, processing, completed, failed
	TxSignature  string    `gorm:"type:varchar(255)"`
	ErrorMessage string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"index"`
	UpdatedAt    time.Time

	Validator *Validator `gorm:"foreignKey:ValidatorID;constraint:OnDelete:CASCADE" json:",omitempty"`
}

func (PayoutTransaction) TableName() string {
	return "PayoutTransaction"
}
