package appderr

type Error interface {
	error

	Code() string

	Message() string

	OrigErr() error
}

type BatchError interface {
	error

	Code() string

	Message() string

	OrigErrs() []error
}

type BatchedErrors interface {
	Error

	OrigErrs() []error
}

func New(code, message string, origErr error) Error {
	var errs []error
	if origErr != nil {
		errs = append(errs, origErr)
	}
	return newBaseError(code, message, errs)
}

func NewBatchError(code, message string, errs []error) BatchedErrors {
	return newBaseError(code, message, errs)
}

type RequestFailure interface {
	Error

	StatusCode() int

	RequestID() string
}

func NewRequestFailure(err Error, statusCode int, reqID string) RequestFailure {
	return newRequestError(err, statusCode, reqID)
}