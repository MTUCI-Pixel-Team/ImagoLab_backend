package user

import (
	"RestAPI/core"
	"RestAPI/db"
	"encoding/json"
	"log"
	"strings"
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

	// if request.FormData.Files != nil {
	// 	for _, fileData := range request.FormData.Files["images"] {
	// 		er := img.SaveFile(fileData.FileName, fileData.FileData)
	// 		if er != nil {
	// 			log.Println("Error saving file:", er)
	// 			return *core.HTTP500.Copy()
	// 		}
	// 	}
	// }

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
		return *core.HTTP500.Copy()
	}
	if !CheckPassword(user.PasswordHash, reqUser.PasswordHash) {
		return *core.HTTP401.Copy()
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
		UserID:           user.ID,
		AccessTokenHash:  accessToken,
		RefreshTokenHash: refreshToken,
	}

	result = db.DB.Save(tokens)
	if result.Error != nil {
		log.Println("Error creating token:", result.Error)
		return *core.HTTP500.Copy()
	}
	user.IsActive = true
	result = db.DB.Save(user)
	if result.Error != nil {
		log.Println("Error saving user:", result.Error)
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

	user.PasswordHash = ""
	user.Tokens = nil

	response := core.HTTP200.Copy()
	err := response.Serialize(user)
	if err != nil {
		log.Println("Error serializing user:", err)
		return *core.HTTP500.Copy()
	}
	return *response
}
