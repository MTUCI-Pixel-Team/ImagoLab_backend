package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
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
func GenerateMigrationFile(structures ...interface{}) error {
	var sb strings.Builder
	sb.WriteString("package migrations\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"RestAPI/db\"\n")
	sb.WriteString("    \"time\"\n")
	sb.WriteString(")\n\n")

	for _, structure := range structures {
		structName := reflect.TypeOf(structure).Name()
		if structName == "" {
			return fmt.Errorf("некорректное имя структуры")
		}

		// Читаем файл models.go
		fileContent, err := ioutil.ReadFile("../db/models.go")
		if err != nil {
			return fmt.Errorf("не удалось прочитать файл: %v", err)
		}

		// Преобразуем содержимое файла в строки
		lines := strings.Split(string(fileContent), "\n")

		// Ищем блок структуры
		var structBlock []string
		inStruct := false
		for _, line := range lines {
			// Начало определения структуры
			if strings.HasPrefix(line, "type "+structName+" struct") {
				inStruct = true
			}

			// Если внутри блока структуры, добавляем строку
			if inStruct {
				structBlock = append(structBlock, line)
			}

			// Конец блока структуры
			if inStruct && line == "}" {
				break
			}
		}

		// Проверяем, найден ли блок структуры
		if len(structBlock) == 0 {
			return fmt.Errorf("определение структуры %s не найдено", structName)
		}
		sb.WriteString(strings.Join(structBlock, "\n"))
	}
	fileText := sb.String()
	file, err := os.Create("migrations/migrate1.go")
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

func main() {
	// Пример вызова функции
	userStruct := GenerateMigrationFile(Token{}, User{})
	fmt.Println(userStruct)
}
