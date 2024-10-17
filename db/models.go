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
	Image{},
}

// Create our models here
type User struct {
	gorm.Model
	Username     string  `json:"username" gorm:"size:64;not null"`
	IsActive     bool    `json:"is_active" gorm:"default:false"`
	Email        string  `json:"email" gorm:"size:256;not null;unique"`
	PasswordHash string  `json:"password,omitempty" gorm:"size:256;not null"`
	Tokens       *Token  `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Image        []Image `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Token struct {
	gorm.Model
	UserID           uint
	AccessTokenHash  string `gorm:"size:256"`
	RefreshTokenHash string `gorm:"size:256"`
	ExpiredAt        time.Time
}

type Image struct {
	gorm.Model
	UserID   uint
	ImageURL string
}
