package godata

type GoDataError struct {
	ResponseCode int
	Message      string
}

func (err *GoDataError) Error() string {
	return err.Message
}

func BadRequestError(message string) *GoDataError {
	return &GoDataError{400, message}
}

func NotFoundError(message string) *GoDataError {
	return &GoDataError{404, message}
}

func MethodNotAllowedError(message string) *GoDataError {
	return &GoDataError{405, message}
}

func GoneError(message string) *GoDataError {
	return &GoDataError{410, message}
}

func PreconditionFailedError(message string) *GoDataError {
	return &GoDataError{412, message}
}

func InternalServerError(message string) *GoDataError {
	return &GoDataError{500, message}
}

func NotImplementedError(message string) *GoDataError {
	return &GoDataError{501, message}
}
