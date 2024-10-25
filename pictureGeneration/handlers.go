package pictureGeneration

import (
	"RestAPI/core"
	"RestAPI/db"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	pg "github.com/prorok210/WS_Client-for_runware.ai-"
)

type User struct {
	db.User
}

var connectedClients = make(map[uint]*pg.WSClient)

/*
docs(

	name: GenerateImage;
	tag: image;
	path: /image/generate;
	method: POST;
	summary: Generate an image using the given data;
	isAuth: true;
	req_content_type:application/json;
	requestbody: {
		"taskType": "string",
		"taskUUID": "string",
		"outputType": ["string"],
		"outputFormat": "string",
		"positivePrompt": "string",
		"negativePrompt": "string",
		"height": int,
		"width": int,
		"model": "string",
		"steps": int,
		"CFGScale": float64,
		"numberResults": int,
		"scheduler": "string",
		"seed": int
	};
	resp_content_type: application/json;
	responsebody: [{
	    "ID": int,
	    "CreatedAt": "time",
	    "UpdatedAt": "time",
	    "DeletedAt": time,
	    "UserID": int,
	    "url": "string"
	}];

)docs
*/
func GenerateImageHandler(request core.HttpRequest) core.HttpResponse {
	if request.User == nil {
		return *core.HTTP401.Copy()
	}
	if request.Method != "POST" {
		return *core.HTTP405.Copy()
	}

	user := request.User.(*db.User)

	newReq := new(pg.ReqMessage)

	err := json.Unmarshal([]byte(request.Body), newReq)
	if err != nil {
		log.Println("Error unmarshalling request body:", err)
		return *core.HTTP400.Copy()
	}

	newReq.TaskUUID = pg.GenerateUUID()

	if connectedClients[user.ID] == nil {
		connectedClients[user.ID] = pg.CreateWsClient(core.RUNWARE_API_KEY, user.ID)
	}

	client := connectedClients[user.ID]

	resp, err := client.SendAndReceiveMsg(*newReq)
	if err != nil {
		log.Println("Error sending request to runware.ai:", err)
		response := core.HTTP500.Copy()
		response.Body = fmt.Sprintf(`{"message":"%s"}`, err.Error())
		return *response
	}

	if len(resp) == 0 {
		return *core.HTTP204.Copy()
	}

	if resp[0].Err != nil {
		log.Println("Error from runware.ai:", resp[0].Err[0].Message)
		response := core.HTTP500.Copy()
		response.Body = fmt.Sprintf(`{"message":"%s"}`, resp[0].Err[0].Message)
		return *response
	}

	imagesData := []db.Image{}

	if len(resp[0].Data) == 0 {
		return *core.HTTP204.Copy()
	}

	for _, respData := range resp {
		for _, data := range respData.Data {
			newImage := db.Image{
				UserID: user.ID,
				Url:    data.ImageURL,
			}
			imagesData = append(imagesData, newImage)
		}
	}

	result := db.DB.Create(&imagesData)
	if result.Error != nil {
		log.Println("Error saving image to database:", result.Error)
		return *core.HTTP500.Copy()
	}
	for _, data := range imagesData {
		user.Images = append(user.Images, data)
	}

	response := core.HTTP201.Copy()
	response.Serialize(user.Images)
	return *response
}

/*
docs(

	name: GetImages;
	tag: image;
	path: /image/get;
	method: GET;
	summary: Get all images of the user;
	isAuth: true;
	QueryParams: {
		"page": "int",
		"limit": "int"
	};
	resp_content_type: application/json;
	responsebody: {
		"total": int,
		"total_pages": int,
		"current_page": int,
		"items": [{
			"ID": int,
			"CreatedAt": "time",
			"UpdatedAt": "time",
			"DeletedAt": time,
			"UserID": int,
			"url": "string"
		}]
	};

)docs
*/
func GetImagesHandler(request core.HttpRequest) core.HttpResponse {
	if request.User == nil {
		return *core.HTTP401.Copy()
	}
	if request.Method != "GET" {
		return *core.HTTP405.Copy()
	}

	user := request.User.(*db.User)

	page := 1
	limit := 10

	if pageStr := request.Query["page"]; pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := request.Query["limit"]; limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var total int64
	if err := db.DB.Model(&db.Image{}).Where("user_id = ?", user.ID).Count(&total).Error; err != nil {
		log.Println("Error counting images:", err)
		return *core.HTTP500.Copy()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	offset := (page - 1) * limit
	images := []db.Image{}
	result := db.DB.Where("user_id = ?", user.ID).
		Offset(offset).
		Limit(limit).
		Find(&images)

	if result.Error != nil {
		log.Println("Error getting images from database:", result.Error)
		return *core.HTTP500.Copy()
	}

	response := core.HTTP200.Copy()
	paginatedResponse := struct {
		Total       int64      `json:"total"`
		TotalPages  int        `json:"total_pages"`
		CurrentPage int        `json:"current_page"`
		Items       []db.Image `json:"items"`
	}{
		Total:       total,
		TotalPages:  totalPages,
		CurrentPage: page,
		Items:       images,
	}

	response.Serialize(paginatedResponse)
	return *response
}
