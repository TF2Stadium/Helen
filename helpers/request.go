package helpers

type Request struct {
	Req  string `json:"request"`
	Data string `json:"data"`
}

func NewRequest(request string, data string) Request {
	return Request{request, data}
}

type RequestObj struct {
	Req  string      `json:"request"`
	Data interface{} `json:"data"`
}

func NewRequestFromObj(request string, data interface{}) RequestObj {
	return RequestObj{request, data}
}
