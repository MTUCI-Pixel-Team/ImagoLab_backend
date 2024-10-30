package db

import (
	"RestAPI/core"
	"fmt"
	"log"
	"strings"

	"RestAPI/migrations"

	"gorm.io/gorm"
)

func migrate(rollback bool, version int) error {
	if rollback {
		return rollbackMigrations(version)
	}
	return applyMigrations()
}

func applyMigrations() error {
	err := migrations.GenerateMigrationFile()
	if err != nil {
		if err.Error() == "no need migrations" {
			return nil
		}
		return fmt.Errorf("failed to generate migration file: %w", err)
	}
	for _, model := range AutoMigrateModels {
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

func rollbackMigrations(version int) error {
	err := migrations.MigrateModelFromVersion(version)
	if err.Error() == "no need migrations" {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

func DropTableMigrations() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		log.Println("Rolling back migrations ...")
		for _, model := range AutoMigrateModels {
			if err := tx.Migrator().DropTable(model); err != nil {
				return fmt.Errorf("failed to drop table for %T: %w", model, err)
			}
			log.Printf("Dropped table for %T", model)
		}
		log.Println("Rolled back migrations successfully")
		return nil
	})
}

func RunMigrations(rollback bool, version int) {
	err := core.InitEnv()
	if err != nil {
		log.Fatalf("Error initializing environment: %v", err)
	}

	err = ConnectToDB(core.DB_CREDENTIALS)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	err = migrate(rollback, version)
	if err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}
}
