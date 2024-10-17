package core

import "strconv"

/*
Стандартные HTTP-ответы
*/
var (
	HTTP200 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Status": "OK"}`)),
		},
		Body: `{"Status": "OK"}`,
	}

	HTTP201 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  201,
		Reason:  "Created",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Status": Created"}`)),
		},
		Body: `{"Status": Created"}`,
	}
	HTTP202 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  202,
		Reason:  "Accepted",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Status": "Accepted"}`)),
		},
		Body: `{"Status": "Accepted"}`,
	}
	HTTP204 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  204,
		Reason:  "No Content",
	}
	HTTP400 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  400,
		Reason:  "Bad Request",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Bad Request"}`)),
		},
		Body: `{"Message": "Bad Request"}`,
	}
	HTTP401 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  401,
		Reason:  "Unauthorized",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Unauthorized"}`)),
		},
		Body: `{"Message": "Unauthorized"}`,
	}
	HTTP403 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  403,
		Reason:  "Forbidden",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Forbidden"}`)),
		},
		Body: `{"Message": "Forbidden"}`,
	}
	HTTP404 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  404,
		Reason:  "Not Found",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Not Found"}`)),
		},
		Body: `{"Message": "Not Found"}`,
	}
	HTTP405 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  405,
		Reason:  "Method Not Allowed",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Method Not Allowed"}`)),
		},
		Body: `{"Message": "Method Not Allowed"}`,
	}
	HTTP408 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  408,
		Reason:  "Request Timeout",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Request Timeout"}`)),
		},
		Body: `{"Message": "Request Timeout"}`,
	}
	HTTP409 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  409,
		Reason:  "Conflict",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Conflict"}`)),
		},
		Body: `{"Message": "Conflict"}`,
	}
	HTTP411 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  411,
		Reason:  "Length Required",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Length Required"}`)),
		},
		Body: `{"Message": "Length Required"}`,
	}
	HTTP415 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  415,
		Reason:  "Unsupported Media Type",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Unsupported Media Type"}`)),
		},
		Body: `{"Message": "Unsupported Media Type"}`,
	}
	HTTP500 = HttpResponse{
		Version: "HTTP/1.1",
		Status:  500,
		Reason:  "Internal Server Error",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": strconv.Itoa(len(`{"Message": "Internal Server Error"}`)),
		},
		Body: `{"Message": "Internal Server Error"}`,
	}
)
