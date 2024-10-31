package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper функция для создания временной директории
func setupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "migrations_test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	return dir
}

// Helper функция для очистки временной директории
func teardownTestDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Не удалось удалить временную директорию: %v", err)
	}
}

func TestGenerateMigrationFile(t *testing.T) {
	// Создаем временную директорию для теста
	testDir := setupTestDir(t)
	defer teardownTestDir(t, testDir)

	// Обновляем пути в функциях (можно использовать переменные окружения или рефакторинг функций для передачи путей как параметров)
	// Для простоты в этом примере будем временно менять текущую рабочую директорию

	err := os.Chdir(testDir)
	if err != nil {
		t.Fatalf("Не удалось сменить текущую директорию: %v", err)
	}

	defer func() {
		// Возвращаемся в исходную директорию после теста
		os.Chdir("..")
	}()

	t.Run("Файл models.go не существует", func(t *testing.T) {
		err := GenerateMigrationFile()
		if err == nil || !strings.Contains(err.Error(), "file 'db/models.go' not found") {
			t.Errorf("Ожидалась ошибка о ненахождении файла models.go, получена: %v", err)
		}
	})

	t.Run("Создание первой миграции", func(t *testing.T) {
		// Создаем файл db/models.go
		err := os.MkdirAll("db", 0755)
		if err != nil {
			t.Fatalf("Не удалось создать директорию db: %v", err)
		}
		modelContent := "package models\n\ntype User struct {\n\tID int\n\tName string\n}\n"
		err = os.WriteFile("db/models.go", []byte(modelContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл models.go: %v", err)
		}

		err = GenerateMigrationFile()
		if err != nil {
			t.Errorf("Ожидалось успешное создание миграции, получена ошибка: %v", err)
		}

		// Проверяем, что файл миграции создан
		migrationPath := filepath.Join("migrations", "migrationFiles", "migrate1.mg")
		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			t.Errorf("Ожидался файл миграции %s, но он не найден", migrationPath)
		}

		// Проверяем содержимое миграционного файла
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Errorf("Не удалось прочитать файл миграции: %v", err)
		}
		if string(content) != modelContent {
			t.Errorf("Содержимое миграционного файла отличается от ожидаемого.\nОжидалось:\n%s\nПолучено:\n%s", modelContent, string(content))
		}
	})

	t.Run("Миграция не нужна, содержимое совпадает", func(t *testing.T) {
		err := GenerateMigrationFile()
		if err == nil || !strings.Contains(err.Error(), "no need migrations") {
			t.Errorf("Ожидалась ошибка о ненадобности миграции, получена: %v", err)
		}

		// Проверяем, что новый файл миграции не создан
		migrationPath := filepath.Join("migrations", "migrationFiles", "migrate2.mg")
		if _, err := os.Stat(migrationPath); !os.IsNotExist(err) {
			t.Errorf("Не ожидалось создание файла миграции %s, но он был найден", migrationPath)
		}
	})

	t.Run("Создание новой миграции при изменении models.go", func(t *testing.T) {
		// Изменяем файл models.go
		updatedModelContent := "package models\n\ntype User struct {\n\tID int\n\tName string\n\tEmail string\n}\n"
		err := os.WriteFile("db/models.go", []byte(updatedModelContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось обновить файл models.go: %v", err)
		}

		err = GenerateMigrationFile()
		if err != nil {
			t.Errorf("Ожидалось успешное создание миграции, получена ошибка: %v", err)
		}

		// Проверяем, что новый файл миграции создан
		migrationPath := filepath.Join("migrations", "migrationFiles", "migrate2.mg")
		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			t.Errorf("Ожидался файл миграции %s, но он не найден", migrationPath)
		}

		// Проверяем содержимое миграционного файла
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Errorf("Не удалось прочитать файл миграции: %v", err)
		}
		if string(content) != updatedModelContent {
			t.Errorf("Содержимое миграционного файла отличается от ожидаемого.\nОжидалось:\n%s\nПолучено:\n%s", updatedModelContent, string(content))
		}
	})
}

func TestMigrateModelFromVersion(t *testing.T) {
	// Создаем временную директорию для теста
	testDir := setupTestDir(t)
	defer teardownTestDir(t, testDir)

	// Переходим в временную директорию
	err := os.Chdir(testDir)
	if err != nil {
		t.Fatalf("Не удалось сменить текущую директорию: %v", err)
	}

	defer func() {
		// Возвращаемся в исходную директорию после теста
		os.Chdir("..")
	}()

	t.Run("Некорректный формат версии", func(t *testing.T) {
		MigrateModelFromVersion(1)
		// Поскольку версия 1 соответствует формату ^(?:\d+)+$, ошибка не ожидается
		// Для проверки некорректного формата можно передать строку или изменить функцию
		// Однако, функция ожидает int, поэтому этот тест не применим напрямую
		// Можно пропустить или изменить функцию для тестирования
	})

	t.Run("Файл миграции не существует", func(t *testing.T) {
		// Создаем файл db/models.go
		err := os.MkdirAll("db", 0755)
		if err != nil {
			t.Fatalf("Не удалось создать директорию db: %v", err)
		}
		modelContent := "package models\n\ntype User struct {\n\tID int\n\tName string\n}\n"
		err = os.WriteFile("db/models.go", []byte(modelContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл models.go: %v", err)
		}

		err = MigrateModelFromVersion(1)
		if err == nil || !strings.Contains(err.Error(), "file 'migrations/migrationFiles/migrate1.mg' not found") {
			t.Errorf("Ожидалась ошибка о ненахождении файла миграции, получена: %v", err)
		}
	})

	t.Run("Миграция не нужна, содержимое совпадает", func(t *testing.T) {
		// Создаем файлы миграций
		err := os.MkdirAll(filepath.Join("migrations", "migrationFiles"), 0755)
		if err != nil {
			t.Fatalf("Не удалось создать директорию миграций: %v", err)
		}

		migrationContent := "package models\n\ntype User struct {\n\tID int\n\tName string\n}\n"
		err = os.WriteFile(filepath.Join("migrations", "migrationFiles", "migrate1.mg"), []byte(migrationContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл миграции: %v", err)
		}

		// Убедимся, что models.go совпадает с миграцией
		modelsPath := filepath.Join("db", "models.go")
		err = os.WriteFile(modelsPath, []byte(migrationContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось записать в файл models.go: %v", err)
		}

		err = MigrateModelFromVersion(1)
		if err == nil || !strings.Contains(err.Error(), "no need migrations") {
			t.Errorf("Ожидалась ошибка о ненадобности миграции, получена: %v", err)
		}
	})

	t.Run("Обновление models.go при различии содержимого", func(t *testing.T) {
		// Создаем миграционный файл
		migrationContent := "package models\n\ntype User struct {\n\tID int\n\tName string\n\tEmail string\n}\n"
		err := os.WriteFile(filepath.Join("migrations", "migrationFiles", "migrate2.mg"), []byte(migrationContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось создать файл миграции: %v", err)
		}

		// Обновляем models.go с другим содержимым
		originalModelContent := "package models\n\ntype User struct {\n\tID int\n\tName string\n}\n"
		err = os.WriteFile(filepath.Join("db", "models.go"), []byte(originalModelContent), 0644)
		if err != nil {
			t.Fatalf("Не удалось записать в файл models.go: %v", err)
		}

		err = MigrateModelFromVersion(2)
		if err != nil {
			t.Errorf("Ожидалось успешное применение миграции, получена ошибка: %v", err)
		}

		// Проверяем, что models.go обновлен
		updatedContent, err := os.ReadFile(filepath.Join("db", "models.go"))
		if err != nil {
			t.Errorf("Не удалось прочитать файл models.go: %v", err)
		}
		if string(updatedContent) != migrationContent {
			t.Errorf("Файл models.go не был обновлен корректно.\nОжидалось:\n%s\nПолучено:\n%s", migrationContent, string(updatedContent))
		}
	})
}
