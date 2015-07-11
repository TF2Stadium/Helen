package helpers

type TPError struct {
	Str  string
	Code int
}

func (e *TPError) Error() string {
	return e.Str
}

func NewTPError(str string, code int) error {
	return &TPError{
		Str:  str,
		Code: code}
}
