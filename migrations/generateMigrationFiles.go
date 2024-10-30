package migrations

import (
	"fmt"
	"io"
	log "log/slog"
	"os"
	"regexp"
	"strconv"
)

/*Функция для генерации txt файла со всеми структурами из db/models.go*/
func GenerateMigrationFile() error {

	filename := "db/models.go"

	// Открываем исходный файл
	srcFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("file '%s' not found: %v", filename, err)
	}
	defer srcFile.Close()

	dirPath := "migrations/migrationFiles"
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	migrationNumber, err := getMaxMigrateNumber()
	if err != nil {
		return fmt.Errorf("failed to get last migration index: %v", err)
	}

	pathToCreateFile := fmt.Sprintf("migrations/migrationFiles/migrate%d.txt", migrationNumber+1)
	lastMigrationFile := fmt.Sprintf("migrations/migrationFiles/migrate%d.txt", migrationNumber)

	// Читаем содержимое последнего файла миграции
	lastFileContent, err := os.ReadFile(lastMigrationFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read last migration file: %v", err)
	}

	// Читаем содержимое нового файла миграции
	newFileContent, err := io.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read new migration file: %v", err)
	}

	// Сравниваем содержимое файлов
	if string(lastFileContent) == string(newFileContent) {
		log.Info("\033[32mnew migration is identical to the last one, no migration needed\033[0m")
		return fmt.Errorf("no need migrations")
	}
	fmt.Println("Generating migration file...")
	// Создаем новый файл миграции
	file, err := os.Create(pathToCreateFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Записываем содержимое в новый файл
	_, err = file.Write(newFileContent)
	if err != nil {
		return fmt.Errorf("error writing content: %v", err)
	}

	log.Info("file created successfully: " + pathToCreateFile)
	return nil
}

func getMaxMigrateNumber() (int, error) {
	files, err := os.ReadDir("migrations/migrationFiles")
	if err != nil {
		return -1, err
	}

	maxNumber := 0
	regex := regexp.MustCompile(`^migrate(\d+)\.txt$`)
	for _, file := range files {
		if !file.IsDir() {
			matches := regex.FindStringSubmatch(file.Name())
			if matches != nil {
				number, err := strconv.Atoi(matches[1])
				if err == nil && number > maxNumber {
					maxNumber = number
				}
			}
		}
	}

	return maxNumber, nil
}
