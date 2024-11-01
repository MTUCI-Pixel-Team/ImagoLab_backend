package core

/*
	Настройки сервера (и приложения)
*/

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	"pictureGeneration",
}

var (
	// Настройки сервера
	PRODUCTION    bool          = false
	HOST          string        = "0.0.0.0"
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
Настройки сертификатов
*/
var (
	CERT_FILE string
	KEY_FILE  string
)

/*
Шаблоны
*/
var (
	MAIL_TEMPLATES_PATH     string
	ACTIVATE_EMAIL_TEMPLATE *template.Template
	RESET_PASSWORD_TEMPLATE *template.Template
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
Ranware settings
*/
var (
	RUNWARE_API_KEY string
)

/*
инициализация переменных окружения
*/
func InitEnv(paths ...string) error {
	var err error
	if !PRODUCTION {
		if len(paths) > 0 {
			err = godotenv.Load(paths...)
		} else {
			err = godotenv.Load()
		}
		if err != nil {
			log.Fatalf("Error env load %v", err)
			return err
		}
	}

	JWT_ACCESS_SECRET_KEY = os.Getenv("JWT_ACCESS_SECRET_KEY")
	JWT_REFRESH_SECRET_KEY = os.Getenv("JWT_REFRESH_SECRET_KEY")
	if JWT_ACCESS_SECRET_KEY == "" || JWT_REFRESH_SECRET_KEY == "" {
		err = errors.New("JWT keys not found")
		log.Fatalf("Error env load %v", err)
		return err
	}

	DB_CREDENTIALS.Password = os.Getenv("DB_PASSWORD")
	if os.Getenv("DB_USER") != "" {
		DB_CREDENTIALS.User = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_HOST") != "" {
		DB_CREDENTIALS.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_NAME") != "" {
		DB_CREDENTIALS.DB_Name = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_PORT") != "" {
		DB_CREDENTIALS.Port, err = strconv.Atoi(os.Getenv("DB_PORT"))
		if err != nil {
			log.Fatalf("Error env load %v", err)
			return err
		}
	}

	if DB_CREDENTIALS.Password == "" {
		err = errors.New("DB password not found")
		log.Fatalf("Error env load %v", err)
		return err
	}

	CERT_FILE = os.Getenv("SSL_CERT_PATH")
	KEY_FILE = os.Getenv("SSL_KEY_PATH")
	if CERT_FILE == "" || KEY_FILE == "" {
		err = errors.New("Cert or key file not found")
		log.Fatalf("Error env load %v", err)
		return err
	}

	MAIL_PASSWORD = os.Getenv("MAIL_PASSWORD")
	if MAIL_PASSWORD == "" {
		err = errors.New("Mail password not found")
		log.Fatalf("Error env load %v", err)
		return err
	}

	MAIL_TEMPLATES_PATH = os.Getenv("MAIL_TEMPLATES_PATH")
	if MAIL_TEMPLATES_PATH == "" {
		err = errors.New("Mail templates path not found")
		log.Fatalf("Error env load %v", err)
		return err
	}

	ACTIVATE_EMAIL_TEMPLATE, err = template.ParseFiles(filepath.Join(MAIL_TEMPLATES_PATH, "Activate.html"))
	RESET_PASSWORD_TEMPLATE, err = template.ParseFiles(filepath.Join(MAIL_TEMPLATES_PATH, "ResetPass.html"))
	if err != nil {
		err = errors.New("Error parsing mail templates")
		log.Fatalf("Error env load %v", err)
		return err
	}

	RUNWARE_API_KEY = os.Getenv("RUNWARE_API_KEY")
	if RUNWARE_API_KEY == "" {
		err = errors.New("Runware API key not found")
		log.Fatalf("Error env load %v", err)
		return err
	}

	if os.Getenv("HTTPS_PORT") != "" {
		HTTPS_PORT, err = strconv.Atoi(os.Getenv("HTTPS_POST"))
		if err != nil {
			log.Fatalf("Error env load %v", err)
			return err
		}
	}

	if os.Getenv("HTTP_PORT") != "" {
		HTTPS_PORT, err = strconv.Atoi(os.Getenv("HTTP_POST"))
		if err != nil {
			log.Fatalf("Error env load %v", err)
			return err
		}
	}

	if os.Getenv("AVATARS_DIR") != "" {
		AVATARS_DIR = os.Getenv("AVATARS_DIR")
	}

	return nil
}
