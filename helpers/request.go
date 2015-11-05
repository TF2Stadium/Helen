package helpers

type Request struct {
	Req  string `json:"request"`
	Data string `json:"data"`
}

func NewRequest(request string, data string) Request {
	return Request{request, data}
}
