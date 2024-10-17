package db

import (
	"RestAPI/core"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB(connData core.DBCredentials) error {
	log.Println("Connecting to database...")
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Moscow",
		connData.User, connData.Password, connData.DB_Name, connData.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to database, error:", err)
		return err
	}
	log.Println("Connected to database successfully")

	DB = db
	return nil
}
