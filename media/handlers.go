package media

import (
	"RestAPI/core"
	"log"
	"os"
)

func ImageHandler(request core.HttpRequest) core.HttpResponse {
	if request.Method != "GET" {
		return core.HTTP405
	}

	currentDir, er := os.Getwd()
	if er != nil {
		log.Println("Error getting current directory:", er)
		return core.HTTP500
	}

	filename := request.Query["filename"]

	filePath := currentDir + core.IMAGES_DIR + "/" + filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return core.HTTP404
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return core.HTTP500
	}

	response := core.HTTP200
	response.Body = string(fileData)
	response.SetHeader("Content-Type", "image/jpeg")
	return response
}

func SaveFile(filename string, fileData []byte) error {
	currentDir, er := os.Getwd()
	filePath := currentDir + core.IMAGES_DIR + "/" + filename
	if er != nil {
		return er
	}
	if _, err := os.Stat(currentDir + core.IMAGES_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(currentDir+core.IMAGES_DIR, 0755)
		if err != nil {
			return err
		}
	}

	file, er := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if er != nil {
		return er
	}
	defer file.Close()

	_, er = file.Write(fileData)
	if er != nil {
		return er
	}
	return nil
}
