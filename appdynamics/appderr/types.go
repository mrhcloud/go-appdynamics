package appderr

import "fmt"

func SprintError(code, message, extra string, origErr error) string {
	msg := fmt.Sprintf("%s: %s", code, message)
	if extra != "" {
		msg = fmt.Sprintf("%s\n\t%s", msg, extra)
	}
	if origErr != nil {
		msg = fmt.Sprintf("%s\ncaused by: %s", msg, origErr.Error())
	}
	return msg
}

type baseError struct {
	code string

	message string

	errs []error
}

func newBaseError(code, message string, origErrs []error) *baseError {
	b := &baseError{
		code: code,
		message: message,
		errs: origErrs,
	}

	return b
}

func (b baseError) Error() string {
	size := len(b.errs)
	if size > 0 {
		return SprintError(b.code, b.message, "", nil)
	}
}

func (b baseError) String() string {
	return b.Error()
}

func (b baseError) Code() string {
	return b.code
}

func (b baseError) Message() string {
	return b.message
}

func (b baseError) OrigErr() error {
	switch len(b.errs) {
	case 0:
		return nil
	case 1:
		return b.errs[0]
	default:
		if err, ok := b.errs[0].(Error); ok {
			return NewBatchError(err. Code(), err.Message(), b.errs[1:])
		}
		return NewBatchError("BatchedErrors",
			"multiple errors occurred", b.errs)
	}
}

func (b baseError) OrigErrs() []error {
	return b.errs
}

type appdError Error

type requestError struct {
	appdError
	statusCode int
	requestID string
}

func newRequestError(err Error, statusCode int, requestID string) *requestError {
	return &requestError{
		appdError: err,
		statusCode: statusCode,
		requestID: requestID,
	}
}

func (r requestError) Error() string {
	extra := fmt.Sprintf("status code: %d, request id: %s",
		r.statusCode, r.requestID)
	return SprintError(r.Code(), r.Message(), extra, r.OrigErr())
}

func (r requestError) String() string {
	return r.Error()
}

func (r requestError) StatusCode() int {
	return r.statusCode
}

func (r requestError) RequestID() string {
	return r.requestID
}

func (r requestError) OrigErrs() []error {
	if b, ok := r.appdError.(BatchedErrors); ok {
		return b.OrigErrs()
	}
	return []error{r.OrigErr()}
}

type errorList []error

func (e errorList) Error() string {
	msg := ""
	if size := len(e); size > 0 {
		for i := 0; i < size; i++ {
			msg += fmt.Sprintf("%s", e[i].Error())
			if i+1 < size {
				msg += "\n"
			}
		}
	}
	return msg
}