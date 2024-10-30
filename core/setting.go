package core

/*
	Настройки сервера (и приложения)
*/

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
)

type DBCredentials struct {
	Host     string
	User     string
	Password string
	DB_Name  string
	Port     int
}

/*
Список приложений, которые будут докумментироваться
*/
var APPS = []string{
	"user",
}

var (
	CERT_FILE string
	KEY_FILE  string
)

const (
	// Настройки сервера
	HOST          string        = "localhost"
	HTTP_PORT     int           = 8082
	HTTPS_PORT    int           = 8446
	CONN_TIMEOUT  time.Duration = 20
	WRITE_TIMEOUT time.Duration = 20
	BUFSIZE       int           = 5 * 1024 * 1024
	AVATARS_DIR   string        = "/media/images/avatars"

	// Настройки мидлваров
	IS_ALLOWED_HOSTS bool = true
	REQ_MIDDLEWARE   bool = true
	KEEP_ALIVE       bool = true
	// Настройки таймаутов для запросов
	AUTH_TIMEOUT time.Duration = time.Minute * 1
)

/*
Настройки базы данных
*/
var DB_CREDENTIALS = DBCredentials{
	Host:    "localhost",
	User:    "admin",
	DB_Name: "dev",
	Port:    5432,
}

/*
Настройки подключений
*/
var ALLOWED_HOSTS = []string{
	"/*",
}

var ALLOWED_METHODS = []string{
	"OPTIONS",
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
}

var SUPPORTED_MEDIA_TYPES = []string{
	"application/json",
	"application/x-www-form-urlencoded",
	"multipart/form-data",
	"text/plain",
	"image/jpeg",
	"image/png",
}

/*
Настройки JWT
*/
var (
	JWT_ACCESS_SECRET_KEY       string
	JWT_REFRESH_SECRET_KEY      string
	JWT_ACCESS_EXPIRATION_TIME  time.Duration = time.Hour * 24
	JWT_REFRESH_EXPIRATION_TIME time.Duration = time.Hour * 336
)

/*
Шаблоны
*/
var (
	ACTIVATE_EMAIL_TEMPLATE *template.Template
	RESET_PASSWORD_TEMPLATE *template.Template
)

/*
Firebase settings
*/
var (
	FIREBASE_API_KEY string
	RECAPTCHA_KEY    string
	PROJECT_ID       string = "imagolab-1729380577888"
)

/*
MAIL SETTINGS
*/
var (
	MAIL_HOST     string = "mail.hosting.reg.ru"
	MAIL_PORT     int    = 465
	MAIL_USER     string = "main@pixel-team.ru"
	MAIL_PASSWORD string
	OTP_EXP_TIME  time.Duration = time.Minute * 5
	OTP_TIMEOUT   time.Duration = time.Minute * 1
)

/*
инициализация переменных окружения
*/
func InitEnv(paths ...string) error {
	var err error
	if len(paths) > 0 {
		err = godotenv.Load(paths...)
	} else {
		err = godotenv.Load()
	}
	if err != nil {
		log.Fatalf("Error env load %v", err)
		return err
	}
	CERT_FILE = fmt.Sprintf("/home/%s/etc/ssl/certs/dev.crt", os.Getenv("HOME_DIRECT"))
	KEY_FILE = fmt.Sprintf("/home/%s/etc/ssl/private/dev.key", os.Getenv("HOME_DIRECT"))
	JWT_ACCESS_SECRET_KEY = os.Getenv("JWT_ACCESS_SECRET_KEY")
	JWT_REFRESH_SECRET_KEY = os.Getenv("JWT_REFRESH_SECRET_KEY")
	if JWT_ACCESS_SECRET_KEY == "" || JWT_REFRESH_SECRET_KEY == "" {
		log.Fatalf("Error env load %v", err)
		return err
	}

	DB_CREDENTIALS.Password = os.Getenv("DB_PASSWORD")
	if DB_CREDENTIALS.Password == "" {
		log.Fatalf("Error env load %v", err)
		return err
	}

	FIREBASE_API_KEY = os.Getenv("FIREBASE_API_KEY")
	RECAPTCHA_KEY = os.Getenv("RECAPTCHA_KEY")
	if FIREBASE_API_KEY == "" || RECAPTCHA_KEY == "" {
		log.Fatalf("Error env load %v", err)
		return err
	}

	MAIL_PASSWORD = os.Getenv("MAIL_PASSWORD")
	if MAIL_PASSWORD == "" {
		log.Fatalf("Error env load %v", err)
		return err
	}

	ACTIVATE_EMAIL_TEMPLATE, err = template.ParseFiles("user/templates/mail/Activate.html")
	if err != nil {
		log.Fatalf("Error env load %v", err)
		return err
	}
	RESET_PASSWORD_TEMPLATE, err = template.ParseFiles("user/templates/mail/ResetPass.html")
	if err != nil {
		log.Fatalf("Error env load %v", err)
		return err
	}

	return nil
}
