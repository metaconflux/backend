package utils

type ApiError struct {
	Code  int
	Error string
}

func NewApiError(code int, err error) (int, ApiError) {
	return code, ApiError{
		Code:  code,
		Error: err.Error(),
	}
}
