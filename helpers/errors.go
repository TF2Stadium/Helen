package helpers

type TPError struct {
	str  string
	code int
}

func (e *TPError) Error() string {
	return e.str
}

func NewError(str string, code int) error {
	return &TPError{
		str:  str,
		code: code}
}
