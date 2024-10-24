package media

import (
	"RestAPI/core"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func ImageHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "GET" {
		return *core.HTTP405.Copy()
	}
	fmt.Println("ImageHandler", request.Url)

	currentDir, er := os.Getwd()
	if er != nil {
		log.Println("Error getting current directory:", er)
		return *core.HTTP500.Copy()
	}

	filename := strings.TrimPrefix(request.Url, "/images/")

	filePath := currentDir + core.AVATARS_DIR + "/" + filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return *core.HTTP404.Copy()
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return *core.HTTP404.Copy()
		}
		log.Println("Error reading file:", err)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	response.Body = string(fileData)
	response.SetHeader("Content-Type", "image/jpeg")
	return *response
}

func SaveFile(fileData []byte, userID uint) (string, error) {
	currentDir, er := os.Getwd()
	if er != nil {
		return "", er
	}

	filename := generateFileName(userID)
	filePath := currentDir + core.AVATARS_DIR + "/" + filename

	if _, err := os.Stat(currentDir + core.AVATARS_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(currentDir+core.AVATARS_DIR, 0755)
		if err != nil {
			return "", err
		}
	}

	file, er := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if er != nil {
		return "", er
	}
	defer file.Close()

	_, er = file.Write(fileData)
	if er != nil {
		return "", er
	}
	return filename, nil
}

func DeleteFile(filename string) error {
	currentDir, er := os.Getwd()
	if er != nil {
		return er
	}
	filePath := currentDir + core.AVATARS_DIR + "/" + filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}

func generateFileName(userID uint) string {
	id := uuid.New().String()
	return strconv.Itoa(int(userID)) + "_" + id + ".jpg"
}
