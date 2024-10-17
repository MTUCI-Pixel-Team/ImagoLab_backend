package db

import (
	"RestAPI/core"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

func migrate(rollback bool) error {
	if rollback {
		return rollbackMigrations()
	}
	return applyMigrations()
}

func applyMigrations() error {
	for _, model := range autoMigrateModels {
		log.Printf("Migrating model: %T", model)
		if err := migrateModel(DB, model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}
	log.Println("Applied migrations for all models")
	return nil
}

func migrateModel(db *gorm.DB, model any) error {
	err := db.AutoMigrate(model)
	if err != nil {
		if strings.Contains(err.Error(), "constraint") && strings.Contains(err.Error(), "does not exist") {
			log.Printf("Warning: Constraint does not exist. Continuing with migration. Error: %v", err)
			return nil
		}
		return err
	}
	return nil
}

func rollbackMigrations() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		log.Println("Rolling back migrations ...")
		for _, model := range autoMigrateModels {
			if err := tx.Migrator().DropTable(model); err != nil {
				return fmt.Errorf("failed to drop table for %T: %w", model, err)
			}
			log.Printf("Dropped table for %T", model)
		}
		log.Println("Rolled back migrations successfully")
		return nil
	})
}

func RunMigrations(rollback bool) {
	err := core.InitEnv()
	if err != nil {
		log.Fatalf("Error initializing environment: %v", err)
	}

	err = ConnectToDB(core.DB_CREDENTIALS)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	err = migrate(rollback)
	if err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}
}
