package user

import (
	"RestAPI/core"
	"RestAPI/db"
	"encoding/json"
	"fmt"
	"log"
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
		"username": "string",
		"email": "string",
		"password": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"username": "string",
		"email": "string",
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

	user.PasswordHash, err = HashPassword(user.PasswordHash)
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
	user.PasswordHash = ""

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
		"email": "string"
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
		"email": "string",
		"activation_code": int
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

	user := new(User)

	result := db.DB.Where("email = ?", reqData.Email).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}
	fmt.Println("user", user)
	fmt.Println("reqData", reqData)

	if user.IsActive {
		resp := core.HTTP409.Copy()
		resp.Body = `{"Message": "User is already activated"}`
		return *resp
	}
	if user.Otp != reqData.Otp || !user.OtpExpires.Before(time.Now()) {
		resp := core.HTTP409.Copy()
		resp.Body = `{"Message": "Invalid activation code"}`
		return *resp
	}

	user.IsActive = true
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
		"email": "string",
		"password": "string"
	};
	resp_content_type: application/json;
	responsebody: {
		"access": "string",
		"refresh": "string"
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

	if !CheckPassword(user.PasswordHash, reqUser.PasswordHash) {
		resp := core.HTTP401.Copy()
		resp.Body = `{"Message": "Invalid email or password"}`
		return *resp
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

	name: GetUserHandler;
	tag: user;
	path: /user/get;
	method: GET;
	summary: Get user by id;
	description: Get user by id from the database;
	isAuth: true;
	resp_content_type: application/json;
	responsebody: {
		"username": "string",
		"email": "string",
	};

)docs
*/
func GetUserHandler(request core.HttpRequest) core.HttpResponse {
	if request.User != nil {
		fmt.Println(request.User)
	}
	if request.Method != "GET" {
		return *core.HTTP405.Copy()
	}
	user := new(User)

	result := db.DB.Where("id = ?", request.Query["id"]).First(user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
		if strings.Contains(result.Error.Error(), "record not found") {
			return *core.HTTP404.Copy()
		}
		return *core.HTTP500.Copy()
	}

	user.Otp = 0
	user.OtpExpires = nil
	user.PasswordHash = ""
	user.Tokens = nil
	user.Image = nil

	response := core.HTTP200.Copy()
	err := response.Serialize(user)
	if err != nil {
		log.Println("Error serializing user:", err)
		return *core.HTTP500.Copy()
	}
	return *response
}
