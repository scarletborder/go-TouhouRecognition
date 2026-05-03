package logic

import "errors"

var (
	errBadRequest = errors.New("bad request")
	errUpstream   = errors.New("upstream error")
	errInternal   = errors.New("internal error")
)

type serviceError struct {
	kind error
	msg  string
}

func (e serviceError) Error() string {
	return e.msg
}

func (e serviceError) Unwrap() error {
	return e.kind
}

func badRequest(message string) error {
	return serviceError{kind: errBadRequest, msg: message}
}

func upstreamError(message string) error {
	return serviceError{kind: errUpstream, msg: message}
}

func internalError(message string) error {
	return serviceError{kind: errInternal, msg: message}
}

func IsBadRequest(err error) bool {
	return errors.Is(err, errBadRequest)
}

func IsUpstreamError(err error) bool {
	return errors.Is(err, errUpstream)
}
