package user

import (
	"RestAPI/core"
	"RestAPI/db"
	"RestAPI/media"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type User struct {
	db.User
}

/*
docs(

	name: CreateUserHandler;
	tag: user;
	path: /user/create;
	method: POST;
	summary: Create a new user;
	description: Create a new user with the given data and save it to the database;
	isAuth: false;
	req_content_type:application/json;
	requestbody: {
		"username*": "string",
		"email*": "string",
		"password*": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"ID": int,
		"CreatedAt": "time",
		"UpdatedAt": "time",
		"DeletedAt": time,
		"username": "string",
		"is_active": bool,
		"email": "string"
	};

)docs
*/
func CreateUserHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}

	user := new(User)
	err := json.Unmarshal([]byte(request.Body), user)
	if err != nil {
		log.Println("Error unmarshaling user:", err)
		return *core.HTTP400.Copy()
	}

	err = ValidateUser(*user)
	if err != nil {
		log.Println("Error validating user:", err)
		resp := core.HTTP400.Copy()
		resp.Body = fmt.Sprintf(`{"Message": "%s"}`, err.Error())
		return *resp
	}

	user.Password, err = HashPassword(user.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return *core.HTTP500.Copy()
	}
	result := db.DB.Create(user)
	if result.Error != nil {
		log.Println("Error creating user:", result.Error)
		if strings.Contains(result.Error.Error(), `duplicate key value violates unique constraint "uni_users_email"`) {
			resp := core.HTTP409.Copy()
			resp.Body = `{"Message": "User with this email already exists"}`
			return *resp
		}
		return *core.HTTP500.Copy()
	}
	user.Password = ""

	response := core.HTTP201.Copy()
	err = response.Serialize(user)
	if err != nil {
		log.Println("Error serializing user:", err)
		return *core.HTTP500.Copy()
	}
	return *response
}

/*
docs(

	name: SendOtpHandler;
	tag: user;
	path: /user/send_otp;
	method: POST;
	summary: Send activation code;
	description: Send activation code to the user's email;
	isAuth: false;
	req_content_type:application/json;
	requestbody: {
		"email*": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"Message": "string"
	};

)docs
*/
func SendOtpHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}

	reqData := new(User)
	err := json.Unmarshal([]byte(request.Body), reqData)
	if err != nil {
		log.Println("Error unmarshaling user:", err)
		return *core.HTTP400.Copy()
	}
	if reqData.Email == "" {
		return *core.HTTP400.Copy()
	}

	user := new(User)
	result := db.DB.Where("email = ?", reqData.Email).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	if user.IsActive {
		resp := core.HTTP409.Copy()
		resp.Body = `{"Message": "User is already activated"}`
		return *resp
	}

	if user.OtpTimeout != nil && user.OtpTimeout.After(time.Now()) {
		resp := core.HTTP429.Copy()
		resp.Body = `{"Message": "Too many requests"}`
		return *resp
	}

	otp := generateActivationCode()

	if user.OtpExpires == nil {
		user.OtpExpires = new(time.Time)
	}
	user.Otp = otp
	*user.OtpExpires = time.Now().Add(core.OTP_EXP_TIME)

	result = db.DB.Save(user)
	if result.Error != nil {
		log.Println("Error saving user:", result.Error)
		return *core.HTTP500.Copy()
	}

	err = SendActivationEmail(user.Email, otp)
	if err != nil {
		log.Println("Error sending email:", err)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	response.Body = `{"Message": "Activation code sent"}`
	return *response
}

/*
docs(

	name: ActivateAccount;
	tag: user;
	path: /user/activate;
	method: POST;
	summary: Activate user account;
	description: Activate user account by activation code;
	isAuth: false;
	req_content_type: application/json;
	requestbody: {
		"email*": "string",
		"otp*": int
	};
	resp_content_type: application/json;
	responsebody: {
		"Message": "string"
	};

)docs
*/
func ActivateAccountHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}
	reqData := new(User)
	err := json.Unmarshal([]byte(request.Body), reqData)
	if err != nil {
		log.Println("Error unmarshaling user:", err)
		return *core.HTTP400.Copy()
	}
	if reqData.Email == "" {
		return *core.HTTP400.Copy()
	}

	user := new(User)

	result := db.DB.Where("email = ?", reqData.Email).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	if user.IsActive {
		resp := core.HTTP409.Copy()
		resp.Body = `{"Message": "User is already activated"}`
		return *resp
	}

	if user.OtpTimeout != nil {
		if user.OtpTimeout.Before(time.Now()) {
			resp := core.HTTP429.Copy()
			resp.Body = fmt.Sprintf(`{"Message": "Too many requests, timeout:%d seconds"}`, int64(user.OtpTimeout.Sub(time.Now()).Seconds()))
			return *resp
		}
	}

	if user.OtpExpires != nil {
		if user.OtpExpires.After(time.Now()) {
			resp := core.HTTP409.Copy()
			resp.Body = `{"Message": "Activation code expired"}`
			return *resp
		}
	}

	if user.Otp != reqData.Otp {
		if user.OtpTries == 5 {
			user.OtpTimeout = new(time.Time)
			*user.OtpTimeout = time.Now().Add(core.OTP_TIMEOUT)
		} else if user.OtpTries == 7 {
			*user.OtpTimeout = time.Now().Add(core.OTP_TIMEOUT * 5)
		} else if user.OtpTries == 10 {
			*user.OtpTimeout = time.Now().Add(core.OTP_TIMEOUT * 10)
		} else if user.OtpTries%5 == 0 {
			*user.OtpTimeout = time.Now().Add(core.OTP_TIMEOUT * 30)
		}
		user.OtpTries++
		resp := core.HTTP409.Copy()
		resp.Body = `{"Message": "Invalid activation code"}`
		return *resp
	} else {
		user.IsActive = true
		user.Otp = 0
		user.OtpTimeout = nil
		user.OtpExpires = nil
		user.OtpTries = 0
	}

	result = db.DB.Save(user)
	if result.Error != nil {
		log.Println("Error saving user:", result.Error)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	response.Body = `{"Message": "User successfully activated"}`
	return *response
}

/*
docs(

	name: AuthUserHandler;
	tag: user;
	path: /user/auth;
	method: POST;
	сontent_type: application/json;
	summary: Authentification user;
	isAuth: false;
	req_content_types: application/json;
	requestbody: {
		"email*": "string",
		"password*": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"CreatedAt": time,
		"UpdatedAt": time,
		"DeletedAt": time,
		"UserID": int,
		"AccessToken": "string",
		"RefreshToken": "string"
	};

)docs
*/
func AuthUserHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}
	reqUser := new(User)
	err := json.Unmarshal([]byte(request.Body), reqUser)
	if err != nil {
		log.Println("Error unmarshaling user:", err)
		return *core.HTTP400.Copy()
	}
	if reqUser.Email == "" || reqUser.Password == "" {
		return *core.HTTP400.Copy()
	}

	user := new(User)

	result := db.DB.Where("email = ?", reqUser.Email).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			resp := core.HTTP404.Copy()
			resp.Body = `{"Message": "User not found"}`
			return *resp
		}
		return *core.HTTP500.Copy()
	}

	if !user.IsActive {
		resp := core.HTTP401.Copy()
		resp.Body = `{"Message": "User is not activated"}`
		return *resp
	}

	if user.AuthTimeout != nil {
		if user.AuthTimeout.After(time.Now()) {
			resp := core.HTTP429.Copy()
			resp.Body = fmt.Sprintf(`{"Message": "Too many requests, timeout:%d seconds"}`, int64(user.AuthTimeout.Sub(time.Now()).Seconds()))
			return *resp
		}
	}

	if !CheckPassword(user.Password, reqUser.Password) {
		if user.AuthTries == 3 {
			user.AuthTimeout = new(time.Time)
			*user.AuthTimeout = time.Now().Add(core.AUTH_TIMEOUT)
		} else if user.AuthTries == 5 {
			*user.AuthTimeout = time.Now().Add(core.AUTH_TIMEOUT * 5)
		} else if user.AuthTries == 8 {
			*user.AuthTimeout = time.Now().Add(core.AUTH_TIMEOUT * 10)
		} else if user.AuthTries%5 == 0 && user.AuthTries > 10 {
			*user.AuthTimeout = time.Now().Add(core.AUTH_TIMEOUT * 30)
		}
		user.AuthTries++
		resp := core.HTTP401.Copy()
		resp.Body = `{"Message": "Invalid email or password"}`
		return *resp
	} else {
		user.AuthTries = 0
		user.AuthTimeout = nil
	}

	accessToken, err := GenerateAccessToken(user.Username, user.Email)
	if err != nil {
		log.Println("Error generating access token:", err)
		return *core.HTTP500.Copy()
	}
	refreshToken, err := GenerateRefreshToken(user.Username, user.Email)
	if err != nil {
		log.Println("Error generating refresh token:", err)
		return *core.HTTP500.Copy()
	}

	tokens := &db.Token{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	result = db.DB.Save(tokens)
	if result.Error != nil {
		log.Println("Error creating token:", result.Error)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	err = response.Serialize(tokens)
	if err != nil {
		log.Println("Error serializing tokens:", err)
		return *core.HTTP500.Copy()
	}

	return *response
}

/*
docs(

	name: RefreshTokenHandler;
	tag: user;
	path: /user/refresh;
	method: POST;
	сontent_type: application/json;
	summary: Refresh tokens;
	isAuth: false;
	req_content_types: application/json;
	requestbody: {
		"refresh_token*": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"CreatedAt": time,
		"UpdatedAt": time,
		"DeletedAt": time,
		"UserID": int,
		"AccessToken": "string",
		"RefreshToken": "string"
	};

)docs
*/
func RefreshTokenHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}

	reqData := new(db.Token)
	err := json.Unmarshal([]byte(request.Body), reqData)
	if err != nil {
		log.Println("Error unmarshaling token:", err)
		return *core.HTTP400.Copy()
	}

	if reqData.RefreshToken == "" {
		return *core.HTTP400.Copy()
	}

	token := new(db.Token)
	result := db.DB.Where("refresh_token = ?", reqData.RefreshToken).First(token)
	if result.Error != nil {
		log.Println("Error finding token:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	claims, err := ValidateToken(reqData.RefreshToken)
	if err != nil {
		log.Println("Error validating token:", err)
		return *core.HTTP401.Copy()
	}
	username := claims["username"].(string)
	email := claims["email"].(string)

	accessToken, err := GenerateAccessToken(username, email)
	if err != nil {
		log.Println("Error generating access token:", err)
		return *core.HTTP500.Copy()
	}
	refreshToen, err := GenerateRefreshToken(username, email)
	if err != nil {
		log.Println("Error generating refresh token:", err)
		return *core.HTTP500.Copy()
	}

	newTokens := &db.Token{
		UserID:       token.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToen,
	}

	result = db.DB.Save(newTokens)
	if result.Error != nil {
		log.Println("Error saving token:", result.Error)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	err = response.Serialize(newTokens)
	if err != nil {
		log.Println("Error serializing tokens:", err)
		return *core.HTTP500.Copy()
	}

	return *response
}

/*
docs(

	name: GetUserHandler;
	tag: user;
	path: /user/get/{int:ID};
	method: GET;
	summary: Get user by id;
	description: Get user by id from the database;
	isAuth: false;
	resp_content_type: application/json;
	responsebody: {
		"ID": int,
		"CreatedAt": "time",
		"UpdatedAt": "time",
		"DeletedAt": time,
		"username": "string",
		"is_active": bool,
		"email": "string"
	};

)docs
*/
func GetUserHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "GET" {
		return *core.HTTP405.Copy()
	}

	id := strings.TrimPrefix(request.Url, "/user/get/")
	userId, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Error converting id:", err)
		return *core.HTTP400.Copy()
	}

	user := new(User)

	result := db.DB.Where("id = ?", userId).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	user.Email = ""
	user.Otp = 0
	user.OtpExpires = nil
	user.Password = ""
	user.Tokens = nil
	user.ResetToken = ""
	user.ResetExpires = nil

	response := core.HTTP200.Copy()
	err = response.Serialize(user)
	if err != nil {
		log.Println("Error serializing user:", err)
		return *core.HTTP500.Copy()
	}
	return *response
}

/*
docs(

	name: GetMeHandler;
	tag: user;
	path: /user/me;
	method: GET;
	summary: Get user by token;
	description: Get user by token;
	isAuth: true;
	resp_content_type: application/json;
	responsebody: {
		"ID": int,
		"CreatedAt": "time",
		"UpdatedAt": "time",
		"DeletedAt": time,
		"username": "string",
		"is_active": bool,
		"email": "string"
	};

)docs
*/
func GetMeHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "GET" {
		return *core.HTTP405.Copy()
	}
	if request.User == nil {
		return *core.HTTP401.Copy()
	}
	reqUser := request.User.(*db.User)

	reqUser.Otp = 0
	reqUser.OtpExpires = nil
	reqUser.Password = ""
	reqUser.Tokens = nil
	reqUser.ResetToken = ""
	reqUser.ResetExpires = nil

	response := core.HTTP200.Copy()
	err := response.Serialize(reqUser)
	if err != nil {
		log.Println("Error serializing user:", err)
		return *core.HTTP500.Copy()
	}
	return *response
}

/*
docs(

	name: UpdateUserHandler;
	tag: user;
	path: /user/update;
	method: PATCH;
	summary: Update user;
	description: Update user with the given data and save it to the database;
	isAuth: true;
	req_content_type: multipart/form-data;
	requestbody: {
		"username": "string",
		"email": "string",
		"file": "avatar",
		"new_password": "string",
		"old_password": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"ID": int,
		"CreatedAt": "time",
		"UpdatedAt": "time",
		"DeletedAt": time,
		"username": "string",
		"is_active": bool,
		"email": "string"
	};

)docs
*/
func UpdateUserHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "PATCH" {
		return *core.HTTP405.Copy()
	}
	if request.User == nil {
		return *core.HTTP401.Copy()
	}

	reqUser := request.User.(*db.User)

	if request.FormData != nil {
		if request.FormData.Fields["username"] == "" && request.FormData.Fields["email"] == "" && request.FormData.Fields["new_password"] == "" && len(request.FormData.Files["avatar"]) == 0 {
			return *core.HTTP400.Copy()
		}
		if request.FormData.Fields["username"] != "" {
			err := ValidateUsername(request.FormData.Fields["username"], DefaultValidationRules())
			if err != nil {
				log.Println("Error validating username:", err)
				resp := core.HTTP400.Copy()
				resp.Body = fmt.Sprintf(`{"Message": "%s"}`, err.Error())
				return *resp
			}
			reqUser.Username = request.FormData.Fields["username"]
		}
		if request.FormData.Fields["email"] != "" {
			err := ValidateEmail(request.FormData.Fields["email"], DefaultValidationRules())
			if err != nil {
				log.Println("Error validating email:", err)
				resp := core.HTTP400.Copy()
				resp.Body = fmt.Sprintf(`{"Message": "%s"}`, err.Error())
				return *resp
			}
			reqUser.Email = request.FormData.Fields["email"]
		}
		if request.FormData.Fields["new_password"] != "" && request.FormData.Fields["old_password"] != "" {
			if !CheckPassword(reqUser.Password, request.FormData.Fields["old_password"]) {
				return *core.HTTP401.Copy()
			}
			valErr := ValidatePassword(request.FormData.Fields["new_password"], DefaultValidationRules())
			if valErr != nil {
				log.Println("Error validating password:", valErr)
				resp := core.HTTP400.Copy()
				resp.Body = fmt.Sprintf(`{"Message": "%s"}`, valErr.Error())
				return *resp
			}

			newPass, err := HashPassword(request.FormData.Fields["new_password"])
			if err != nil {
				log.Println("Error hashing password:", err)
				return *core.HTTP500.Copy()
			}
			reqUser.Password = newPass
		}
		if len(request.FormData.Files["avatar"]) > 0 {
			filename, err := media.SaveFile(request.FormData.Files["avatar"][0].FileData, reqUser.ID)
			if err != nil {
				log.Println("Error saving file:", err)
				return *core.HTTP500.Copy()
			}
			if reqUser.Avatar != "" {
				err := media.DeleteFile(strings.TrimPrefix(reqUser.Avatar, "/images/"))
				if err != nil {
					log.Println("Error deleting file:", err)
					return *core.HTTP500.Copy()
				}
			}
			reqUser.Avatar = "/images/" + filename
		} else {
			if reqUser.Avatar != "" {
				err := media.DeleteFile(strings.TrimPrefix(reqUser.Avatar, "/images/"))
				if err != nil {
					log.Println("Error deleting file:", err)
					return *core.HTTP500.Copy()
				}
				reqUser.Avatar = ""
			}
		}
	} else {
		return *core.HTTP400.Copy()
	}

	result := db.DB.Save(reqUser)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), `duplicate key value violates unique constraint "uni_users_email"`) {
			resp := core.HTTP409.Copy()
			resp.Body = `{"Message": "User with this email already exists"}`
			return *resp
		}
		log.Println("Error saving user:", result.Error)
		return *core.HTTP500.Copy()
	}

	reqUser.Password = ""
	reqUser.Otp = 0
	reqUser.OtpExpires = nil
	reqUser.Tokens = nil
	reqUser.ResetToken = ""
	reqUser.ResetExpires = nil

	response := core.HTTP200.Copy()
	err := response.Serialize(reqUser)
	if err != nil {
		log.Println("Error serializing user:", err)
		return *core.HTTP500.Copy()
	}
	return *response
}

/*
docs(

	name: SendResetPasswordMailHandler;
	tag: user;
	path: /user/send_reset_password_mail;
	method: POST;
	summary: Send mail for reset password;
	description: Send mail with reset code to the user's email;
	isAuth: false;
	req_content_type:application/json;
	requestbody: {
		"email*": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"Message": "string"
	};

)docs
*/
func SendResetPasswordMailHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}
	reqUser := new(User)
	err := json.Unmarshal([]byte(request.Body), reqUser)
	if err != nil {
		log.Println("Error unmarshaling user:", err)
		return *core.HTTP400.Copy()
	}
	if reqUser.Email == "" {
		return *core.HTTP400.Copy()
	}

	result := db.DB.Where("email = ?", reqUser.Email).First(reqUser)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	if reqUser.ResetTimeout != nil {
		if reqUser.ResetTimeout.After(time.Now()) {
			resp := core.HTTP429.Copy()
			resp.Body = fmt.Sprintf(`{"Message": "Too many requests, timeout:%d seconds"}`, int64(reqUser.ResetTimeout.Sub(time.Now()).Seconds()))
			return *resp
		}
	}

	resetCode, err := generateSecureToken()
	if err != nil {
		log.Println("Error generating reset code:", err)
		return *core.HTTP500.Copy()
	}

	reqUser.ResetToken = resetCode
	reqUser.ResetExpires = new(time.Time)
	*reqUser.ResetExpires = time.Now().Add(core.OTP_EXP_TIME)

	result = db.DB.Save(reqUser)
	if result.Error != nil {
		log.Println("Error saving user:", result.Error)
		return *core.HTTP500.Copy()
	}

	err = SendResetPasswordEmail(reqUser.Email, resetCode)
	if err != nil {
		log.Println("Error sending email:", err)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	response.Body = `{"Message": "Reset code sent"}`
	return *response
}

/*
docs(

	name: ResetPasswordHandler;
	tag: user;
	path: /user/reset_password;
	method: POST;
	summary: Reset password;
	description: Reset password by reset code;
	isAuth: false;
	req_content_type: application/json;
	requestbody: {
		"email*": "string",
		"reset_token*": string,
		"password*": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"Message": "string"
	};

)docs
*/
func ResetPasswordHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}
	reqUser := new(User)
	err := json.Unmarshal([]byte(request.Body), reqUser)
	if err != nil {
		log.Println("Error unmarshaling user:", err)
		return *core.HTTP400.Copy()
	}
	if reqUser.Email == "" || reqUser.ResetToken == "" || reqUser.Password == "" {
		return *core.HTTP400.Copy()
	}

	valErr := ValidatePassword(reqUser.Password, DefaultValidationRules())
	if valErr != nil {
		log.Println("Error validating password:", err)
		resp := core.HTTP400.Copy()
		resp.Body = fmt.Sprintf(`{"Message": "%s"}`, valErr.Error())
		return *resp
	}

	user := new(User)

	result := db.DB.Where("email = ? and reset_token = ?", reqUser.Email, reqUser.ResetToken).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	if user.ResetTimeout != nil {
		if user.ResetTimeout.After(time.Now()) {
			resp := core.HTTP429.Copy()
			resp.Body = fmt.Sprintf(`{"Message": "Too many requests, timeout:%d seconds"}`, int64(user.ResetTimeout.Sub(time.Now()).Seconds()))
			return *resp
		}
	}
	if user.ResetExpires != nil {
		if user.ResetExpires.Before(time.Now()) {
			resp := core.HTTP409.Copy()
			resp.Body = `{"Message": "Reset code expired"}`
			return *resp
		}
	}

	if user.ResetToken != reqUser.ResetToken {
		if user.ResetTries == 5 {
			user.ResetTimeout = new(time.Time)
			*user.ResetTimeout = time.Now().Add(core.OTP_TIMEOUT)
		} else if user.ResetTries == 7 {
			*user.ResetTimeout = time.Now().Add(core.OTP_TIMEOUT * 5)
		} else if user.ResetTries == 10 {
			*user.ResetTimeout = time.Now().Add(core.OTP_TIMEOUT * 10)
		} else if user.ResetTries%5 == 0 && user.ResetTries > 10 {
			*user.ResetTimeout = time.Now().Add(core.OTP_TIMEOUT * 30)
		}
		user.ResetTries++
		resp := core.HTTP409.Copy()
		resp.Body = `{"Message": "Invalid reset code"}`
		return *resp
	} else {
		user.ResetTries = 0
		user.ResetTimeout = nil
		newPass, err := HashPassword(reqUser.Password)
		if err != nil {
			log.Println("Error hashing password:", err)
			return *core.HTTP500.Copy()
		}
		user.Password = newPass
	}

	result = db.DB.Save(user)
	if result.Error != nil {
		log.Println("Error saving user:", result.Error)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	response.Body = `{"Message": "Password successfully reset"}`
	return *response
}
