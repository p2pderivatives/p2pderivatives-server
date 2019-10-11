package database

// A code describing the kind of error that occured.
type Code uint32

const (
	// InternalError indicates that a system error while performing the
	// requested operation.
	InternalError Code = 1
	// NotFound indicates that the requested element was not found in the
	// database.
	NotFound Code = 2
)

// DbError is an error returned by the database layer.
type DbError struct {
	message string
	code    Code
}

// NewDbError creates a new DbError.
func NewDbError(message string, code Code) *DbError {
	return &DbError{message: message, code: code}
}

func (err DbError) Error() string {
	return err.message
}

func (err DbError) Code() Code {
	return err.code
}
