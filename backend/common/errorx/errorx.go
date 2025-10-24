package errorx

// StatCodeError 带状态码的错误
type StatCodeError struct {
	Code    int    `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *StatCodeError) Error() string {
	return e.Message
}

// NewStatCodeError 创建带状态码的错误
func NewStatCodeError(status, code int, message string) *StatCodeError {
	return &StatCodeError{
		Code:    code,
		Status:  status,
		Message: message,
	}
}

// NewDefaultError 创建默认错误
func NewDefaultError(message string) *StatCodeError {
	return &StatCodeError{
		Code:    500,
		Status:  500,
		Message: message,
	}
}

// NewBadRequestError 创建400错误
func NewBadRequestError(message string) *StatCodeError {
	return &StatCodeError{
		Code:    400,
		Status:  400,
		Message: message,
	}
}

// NewNotFoundError 创建404错误
func NewNotFoundError(message string) *StatCodeError {
	return &StatCodeError{
		Code:    404,
		Status:  404,
		Message: message,
	}
}

// NewInternalServerError 创建500错误
func NewInternalServerError(message string) *StatCodeError {
	return &StatCodeError{
		Code:    500,
		Status:  500,
		Message: message,
	}
}
