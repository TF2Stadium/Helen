package helpers

type Request struct {
	Req  string      `json:"request"`
	Data interface{} `json:"data"`
}

func NewRequest(request string, data interface{}) Request {
	return Request{request, data}
}
