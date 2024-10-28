package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Token struct {
	ID           uint `json:"-" gorm:"primaryKey"`
	UserID       uint
	AccessToken  string `json:"access_token,omitempty," gorm:"size:256"`
	RefreshToken string `json:"refresh_token,omitempty" gorm:"size:256"`
}

// Пример структуры, которую мы будем использовать
type User struct {
	Username     string     `json:"username" gorm:"size:64;not null"`
	IsActive     bool       `json:"is_active" gorm:"default:false"`
	Email        string     `json:"email" gorm:"size:256;not null;unique"`
	Password     string     `json:"password,omitempty" gorm:"size:256;not null"`
	Avatar       string     `json:"avatar,omitempty"`
	Otp          int        `json:"otp,omitempty"`
	OtpExpires   *time.Time `json:"otp_expires,omitempty"`
	OtpTries     int        `json:"-" gorm:"default:0"`
	OtpTimeout   *time.Time `json:"-" gorm:"type:timestamp"`
	ResetToken   string     `json:"reset_token,omitempty"`
	ResetExpires *time.Time `json:"-"`
	ResetTries   int        `json:"-" gorm:"default:0"`
	ResetTimeout *time.Time `json:"-" gorm:"type:timestamp"`
	AuthTries    int        `json:"-" gorm:"default:0"`
	AuthTimeout  *time.Time `json:"-" gorm:"type:timestamp"`
	Tokens       *Token     `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Images       []Image    `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
type Image struct{}

// Функция для генерации строки описания структур
func GenerateMigrationFile() error {
	var sb strings.Builder
	sb.WriteString("package migrations\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"gorm.io/gorm\"\n")
	sb.WriteString("    \"time\"\n")
	sb.WriteString(")\n\n")

	// Читаем файл models.go
	fileContent, err := os.ReadFile("../db/models.go")
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %v", err)
	}

	// Преобразуем содержимое файла в строки
	lines := strings.Split(string(fileContent), "\n")

	/*
		Элементы - строки структуры
	*/
	var structBlock []string
	inStruct := false
	isAutoMigrateModels := false

	for _, line := range lines {
		if strings.HasPrefix(line, "var AutoMigrateModels = []any{") {
			isAutoMigrateModels = true
		}

		if isAutoMigrateModels {
			structBlock = append(structBlock, line)
		}

		if isAutoMigrateModels && line == "}" {
			structBlock = append(structBlock, "")
			break
		}

	}

	for _, line := range lines {
		// Начало определения структуры
		if strings.HasPrefix(line, "type") {
			inStruct = true
		}

		// Если внутри блока структуры, добавляем строку
		if inStruct {
			structBlock = append(structBlock, line)
		}

		// Конец блока структуры
		if inStruct && line == "}" {
			structBlock = append(structBlock, "")
		}
	}

	// Проверяем, найден ли блок структуры
	if len(structBlock) == 0 {
		return fmt.Errorf("в файле models нет моделей")
	}
	sb.WriteString(strings.Join(structBlock, "\n"))

	fileText := sb.String()
	migrationNumber, err := getMaxMigrateNumber()

	fmt.Println("migrationNumber", migrationNumber)

	if err != nil {
		return fmt.Errorf("не удалось получить индекс последней миграции: %v", err)
	}
	pathToCreateFile := fmt.Sprintf("migrationFiles/migrate%d.txt", migrationNumber+1)
	file, err := os.Create(pathToCreateFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Запись строки структуры в файл
	_, err = file.WriteString(fileText)
	if err != nil {
		return err
	}
	return nil
}

/*
Функция для получения индекса последней миграции
*/
func getMaxMigrateNumber() (int, error) {
	files, err := ioutil.ReadDir("migrationFiles")
	if err != nil {
		return 0, err
	}

	maxNumber := 0
	regex := regexp.MustCompile(`^migrate(\d+)\.txt$`)
	fmt.Println("files", files)
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

func main() {
	// Пример вызова функции
	userStruct := GenerateMigrationFile()
	fmt.Println(userStruct)
}
