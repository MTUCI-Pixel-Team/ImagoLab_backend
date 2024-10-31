package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"strconv"
	"strings"
)

/*
Структуры для работы с HTTP-запросами и ответами
*/
type HttpRequest struct {
	Method   string
	Url      string
	Query    map[string]string
	Version  string
	Headers  map[string]string
	User     interface{}
	Body     string
	FormData *FormData
}

type FormData struct {
	Fields map[string]string
	Files  map[string][]struct {
		FileName string
		FileData []byte
	}
}

type HttpResponse struct {
	Version string
	Status  int
	Reason  string
	Headers map[string]string
	Body    string
}

/*
	Методы для работы с HTTP-запросами и ответами
	ParseRequest() - разбор HTTP-запроса из байтового массива в структуру HttpRequest, возвращает ошибку в случае некоретного запроса
	ToString() - преобразование HTTP-запроса в строку
	ToBytes() - преобразование HTTP-ответа в байтовый массив для отправки по сети
	SetHeader() - установка заголовка в HTTP-ответе
	ParseFormData() - разбор multipart/form-data из тела HTTP-запроса
	Serialize() - сериализация данных в JSON и запись в тело HTTP-ответа
	Copy() - копирование HTTP-ответа
*/

func (rqst *HttpRequest) ParseRequest(buffer []byte) error {
	if len(buffer) == 0 {
		return errors.New("Empty request")
	}

	reqStr := string(buffer)
	lines := strings.Split(reqStr, "\r\n")
	if len(lines) < 1 {
		return errors.New("Invalid request: no lines found")
	}

	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 3 {
		return errors.New("Invalid request line")
	}
	rqst.Method = requestLine[0]
	UrlAndQuery := strings.Split(requestLine[1], "?")
	rqst.Url = UrlAndQuery[0]
	rqst.Version = requestLine[2]
	if rqst.Method == "" || rqst.Url == "" || rqst.Version == "" {
		return errors.New("Invalid request line")
	}
	rqst.Query = make(map[string]string)
	rqst.Headers = make(map[string]string)

	if len(UrlAndQuery) > 1 {
		queryParts := strings.Split(UrlAndQuery[1], "&")
		for _, queryPart := range queryParts {
			queryParam := strings.Split(queryPart, "=")
			if len(queryParam) == 2 {
				rqst.Query[queryParam[0]] = queryParam[1]
			} else if len(queryParam) == 1 {
				rqst.Query[queryParam[0]] = ""
			}
		}
	}

	i := 1
	for i < len(lines) && lines[i] != "" {
		headerParts := strings.SplitN(lines[i], ": ", 2)
		if len(headerParts) == 2 {
			rqst.Headers[headerParts[0]] = headerParts[1]
		}
		i++
	}

	if i+1 < len(lines) {
		rqst.Body = strings.Join(lines[i+1:], "\r\n")
	} else {
		rqst.Body = ""
	}

	return nil
}

func (rqst *HttpRequest) ToString() string {
	reqStr := rqst.Method + " " + rqst.Url + " " + rqst.Version + "\r\n"
	for key, value := range rqst.Headers {
		reqStr += key + ": " + value + "\r\n"
	}
	reqStr += "\r\n" + rqst.Body

	return reqStr
}

func (resp *HttpResponse) ToString() string {
	respStr := resp.Version + " " + strconv.Itoa(resp.Status) + " " + resp.Reason + "\r\n"
	for key, value := range resp.Headers {
		respStr += key + ": " + value + "\r\n"
	}
	respStr += "\r\n" + resp.Body

	return respStr
}

func (resp *HttpResponse) ToBytes() []byte {
	return []byte(resp.ToString())
}

func (rqst *HttpRequest) ToBytes() []byte {
	return []byte(rqst.ToString())
}

func (resp *HttpResponse) SetHeader(key string, value string) {
	if key == "" || value == "" {
		return
	}
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}
	resp.Headers[key] = value
}

func (req *HttpRequest) ParseFormData() error {
	req.FormData = &FormData{
		Fields: make(map[string]string),
		Files: make(map[string][]struct {
			FileName string
			FileData []byte
		}),
	}

	contentType := req.Headers["Content-Type"]
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "multipart/form-data" {
		return errors.New("invalid Content-Type")
	}

	boundary, ok := params["boundary"]
	if !ok {
		return errors.New("boundary not found in Content-Type")
	}

	reader := multipart.NewReader(bytes.NewReader([]byte(req.Body)), boundary)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := part.FormName()
		if name == "" {
			return errors.New("invalid Content-Disposition: name not found")
		}

		filename := part.FileName()
		if filename != "" {
			fileData, err := io.ReadAll(part)
			if err != nil {
				return err
			}

			req.FormData.Files[name] = append(req.FormData.Files[name], struct {
				FileName string
				FileData []byte
			}{
				FileName: filename,
				FileData: fileData,
			})
		} else {
			fieldValue, err := io.ReadAll(part)
			if err != nil {
				return err
			}
			req.FormData.Fields[name] = string(fieldValue)
		}
	}

	return nil
}

func (resp *HttpResponse) Serialize(data interface{}) error {
	if data == nil {
		return errors.New("Data is nil")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp.Body = string(jsonData)
	return nil
}

func (resp *HttpResponse) Copy() *HttpResponse {
	newResp := &HttpResponse{
		Version: resp.Version,
		Status:  resp.Status,
		Reason:  resp.Reason,
		Headers: make(map[string]string),
		Body:    resp.Body,
	}

	for key, value := range resp.Headers {
		newResp.Headers[key] = value
	}
	return newResp
}
