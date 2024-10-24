package app

import (
	"RestAPI/docs"
	"RestAPI/media"
	"RestAPI/user"
)

/*
	Функция InitHandler() - инициализация списка представлений
	После создания представлений их необходимо зарегистрировать в этой функции, чтобы они были доступны для обработки запросов
	Для регистрации нужно передать url, по которому будет доступно представление, указатель на функцию-обработчик и имя предсталвения(оно должно совпадать с именем в документации для корректной работы)
	При регистрации роута можно использовать плейсхолдеры вида {int:<int>} или {<string>} для передачи параметров в запросе
	Роутер выдаст указатель на функцию, которая будет обрабатывать запрос или nil, если функции не нашлось
*/

func InitHandlers() {
	registerHandler("/api/docs", docs.GetDocs, "docs")
	registerHandler("/api/docs/templates/css/styles.css", docs.GetDocsCSS, "docs")
	registerHandler("/api/docs/templates/js/script.js", docs.GetDocsJS, "docs")
	registerHandler("/images", media.ImageHandler, "images")
	registerHandler("/user/create", user.CreateUserHandler, "createUser")
	registerHandler("/user/send_otp", user.SendOtpHandler, "sendOtp")
	registerHandler("/user/activate", user.ActivateAccountHandler, "activateUser")
	registerHandler("/user/auth", user.AuthUserHandler, "verifyUser")
	registerHandler("/user/get/{int:ID}", user.GetUserHandler, "getUser")
	registerHandler("/user/update", user.UpdateUserHandler, "updateUser")
}
