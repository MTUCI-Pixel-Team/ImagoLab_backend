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

	migrationFilename := "migrations/migrationFiles/migrate" + version + ".txt"

	// Открываем исходный файл
	srcFile, err := os.Open(migrationFilename)
	if err != nil {
		return fmt.Errorf("file '%s' not found: %v", migrationFilename, err)
	}
	defer srcFile.Close()

	// Создаем имя файла с расширением .go
	modelsFilename := "db/models.go"

	// Читаем содержимое последнего файла миграции
	lastFileContent, err := os.ReadFile(modelsFilename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read models file: %v", err)
	}

	// Читаем содержимое нового файла миграции
	newFileContent, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Сравниваем содержимое файлов
	if string(lastFileContent) == string(newFileContent) {
		log.Info("\033[32mmigration is identical to the models file, no migration needed\033[0m")
		return fmt.Errorf("no need migrations")
	}

	// Создаем файл назначения
	dstFile, err := os.Create(modelsFilename)
	if err != nil {
		return fmt.Errorf("error creating file '%s': %v", modelsFilename, err)
	}
	defer dstFile.Close()

	// Копируем содержимое
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	log.Info("The file was successfully copied from " + migrationFilename + " into " + strings.TrimPrefix(modelsFilename, "../"))
	return nil
}
