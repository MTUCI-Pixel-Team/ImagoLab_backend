package db

import (
	"time"

	"gorm.io/gorm"
)

/*
Для автоматической миграции моделей в базу данных внести модель в массив AutoMigrateModels
*/
var autoMigrateModels = []any{
	User{},
	Token{},
}

// Create our models here
type User struct {
	gorm.Model
	Username     string     `json:"username" gorm:"size:64;not null"`
	IsActive     bool       `json:"is_active" gorm:"default:false"`
	Email        string     `json:"email" gorm:"size:256;not null;unique"`
	Password     string     `json:"password,omitempty" gorm:"size:256;not null"`
	Avatar       string     `json:"avatar,omitempty"`
	Otp          int        `json:"otp,omitempty"`
	OtpExpires   *time.Time `json:"otp_expires,omitempty" gorm:"autoUpdateTime"`
	OtpTries     int        `json:"-" gorm:"default:0"`
	OtpTimeout   *time.Time `json:"-" gorm:"type:timestamp"`
	ResetToken   string     `json:"reset_token,omitempty"`
	ResetExpires *time.Time `json:"reset_expires,omitempty" gorm:"autoUpdateTime"`
	ResetTries   int        `json:"-" gorm:"default:0"`
	ResetTimeout *time.Time `json:"-" gorm:"type:timestamp"`
	AuthTries    int        `json:"-" gorm:"default:0"`
	AuthTimeout  *time.Time `json:"-" gorm:"type:timestamp"`

	Tokens *Token `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Token struct {
	ID           uint `json:"-" gorm:"primaryKey"`
	UserID       uint
	AccessToken  string `gorm:"size:256"`
	RefreshToken string `gorm:"size:256"`
}
