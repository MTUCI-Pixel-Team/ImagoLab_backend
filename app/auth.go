package app

import (
	"RestAPI/core"
	"RestAPI/db"
	"RestAPI/user"
	"fmt"
	"strings"
)

func CheckAuth(req *core.HttpRequest) {
	token, ok := req.Headers["Authorization"]
	if !ok || !strings.HasPrefix(token, "Bearer ") {
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")

	tokenRecord := new(db.Token)
	result := db.DB.Where("access_token = ?", token).First(tokenRecord)
	if result.Error != nil {
		return
	}

	userDB := new(db.User)
	result = db.DB.First(userDB, tokenRecord.UserID)
	if result.Error != nil {
		return
	}
	if !userDB.IsActive {
		return
	}

	claims, err := user.ValidateToken(token)
	if err != nil {
		fmt.Println("Error validating token:", err)
		return
	}
	if claims["token_type"] != "access" {
		return
	}
	req.User = userDB
}
