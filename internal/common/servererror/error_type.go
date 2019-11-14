package servererror

// ErrorCode represents an error code.
type ErrorCode int

const (
	// InternalError represents an error caused by an unexpected state.
	InternalError ErrorCode = iota + 1
	// InvalidArguments is returned when arguments passed to the server are
	// invalid.
	InvalidArguments
	// DeadlineExceeded is returned when a timeout occured while processing the
	// request.
	DeadlineExceeded
	// DbError is returned when the database encountered an error.
	DbError
	// NotFoundError is returned when the requested resource was not found on
	// the server.
	NotFoundError
	// AlreadyExistError is returned when trying to create a resource that
	// already exists on the server.
	AlreadyExistError
	// OptimisticLockError is returned when an error occured with the optimistic
	// lock.
	OptimisticLockError
	// PreconditionError is returned when a precondition was violated.
	PreconditionError
	// Unavailable is returned when the server is currently unavailable.
	Unavailable
	// UnauthenticatedError is returned when the requested action requires
	// authentication but the request did not contain authentication information.
	UnauthenticatedError
	// UnknownError is returned when an error occured but the cause of the error
	// is unknown.
	UnknownError
	// PermissionDenied is returned when the requested action cannot be
	// performed given the provided authentication.
	PermissionDenied
)

// Error represent an error in the system.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details []ErrorDetail
}

// Error returns the message associated with the error.
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new Error structure.
func NewError(code ErrorCode, message string, err error) error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
		Details: nil,
	}
}

// ErrorDetailCode indicates a detailed error code.
type ErrorDetailCode int

const (
	// ErrorDetailCodeUnknown indicates an internal error.
	ErrorDetailCodeUnknown ErrorDetailCode = iota + 1

	// ErrorDetailCodeTokenRequired indicates that the requested service
	// required an authentication token that was not provided.
	ErrorDetailCodeTokenRequired
	// ErrorDetailCodeTokenExpired indicates that the requested service
	// required an authentication token but the provided one was expired.
	ErrorDetailCodeTokenExpired
	// ErrorDetailCodeTokenInvalid indicates that the requested service
	// required an authentication token but the provided one was invalid.
	ErrorDetailCodeTokenInvalid
)

// ErrorDetail contains detailed information about an error.
type ErrorDetail struct {
	Code   ErrorDetailCode `json:"code"`   // エラー詳細コード
	Values []string        `json:"values"` // エラー詳細情報
}

// NewErrorWithDetail creates a new ErrorDetail structure containing the
// provided information.
func NewErrorWithDetail(code ErrorCode, message string, err error, detailCode ErrorDetailCode, detailValues []string) error {
	return NewErrorWithDetails(
		code,
		message,
		err,
		[]ErrorDetail{{Code: detailCode, Values: detailValues}},
	)
}

// NewErrorWithDetails creates a new ErrorDetail structure containing the
// provided information.
func NewErrorWithDetails(code ErrorCode, message string, err error, details []ErrorDetail) error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
		Details: details,
	}
}
