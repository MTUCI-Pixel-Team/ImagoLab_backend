package migrations

import (
	"fmt"
	"io"
	log "log/slog"
	"os"
	"regexp"
	"strings"
)

/*Функция для миграции модели из файла, необходимо передать версию миграции*/
func MigrateModelFromVersion(v int) error {
	// Добавляем расширение .txt, если его нет
	version := fmt.Sprintf("%d", v)
	re := regexp.MustCompile(`^(?:\d+)+$`)
	if !re.MatchString(version) {
		return fmt.Errorf("invalid migration version format")
	}

	// Открываем исходный файл
	migrationFilename := "migrations/migrationFiles/migrate" + version + ".mg"
	migrationFile, err := os.Open(migrationFilename)
	if err != nil {
		return fmt.Errorf("file '%s' not found: %v", migrationFilename, err)
	}
	defer migrationFile.Close()

	// Читаем содержимое файла миграции
	migrationFileText, err := io.ReadAll(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Открываем файл с моделями
	modelsFilename := "db/models.go"
	modelsFile, err := os.Open(modelsFilename)
	if err != nil {
		return fmt.Errorf("file '%s' not found: %v", modelsFilename, err)
	}
	defer modelsFile.Close()

	// Читаем содержимое файла моделей
	modelsFileText, err := io.ReadAll(modelsFile)
	if err != nil {
		return fmt.Errorf("failed to read models file: %v", err)
	}

	// Сравниваем содержимое файлов
	if string(migrationFileText) == string(modelsFileText) {
		log.Info("\033[32mmigration is identical to the models file, no migration needed\033[0m")
		return fmt.Errorf("no need migrations")
	}

	goFile, err := os.Create(modelsFilename)
	if err != nil {
		log.Warn("error creating file")
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer goFile.Close()

	// Записываем содержимое .txt в .go
	_, err = goFile.Write(migrationFileText)
	if err != nil {
		log.Warn("error writing to file db/models.go")
		return fmt.Errorf("error writing content: %v", err)
	}

	log.Info("The file was successfully copied from " + migrationFilename + " into " + strings.TrimPrefix(modelsFilename, "../"))
	return nil
}
