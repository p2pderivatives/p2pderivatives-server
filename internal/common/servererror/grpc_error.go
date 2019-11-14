package servererror

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewNotFoundStatus returns a GRPC status with the NotFound code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewNotFoundStatus(message string) *status.Status {
	return status.New(codes.NotFound, message)
}

// NewInternalStatus returns a GRPC status with the Internal code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewInternalStatus(message string) *status.Status {
	return status.New(codes.Internal, message)
}

// NewAlreadyExistsStatus returns a GRPC status with the AlreadyExists code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewAlreadyExistsStatus(message string) *status.Status {
	return status.New(codes.AlreadyExists, message)
}

// NewInvalidArgumentStatus returns a GRPC status with the InvalidArgument code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewInvalidArgumentStatus(message string) *status.Status {
	return status.New(codes.InvalidArgument, message)
}

// NewDeadlineExceededStatus returns a GRPC status with the DeadlineExceeded code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewDeadlineExceededStatus(message string) *status.Status {
	return status.New(codes.DeadlineExceeded, message)
}

// NewFailedPreconditionStatus returns a GRPC status with the FailedPrecondition code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewFailedPreconditionStatus(message string) *status.Status {
	return status.New(codes.FailedPrecondition, message)
}

// NewUnimplementedStatus returns a GRPC status with the Unimplemented code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewUnimplementedStatus(message string) *status.Status {
	return status.New(codes.Unimplemented, message)
}

// NewUnavailableStatus returns a GRPC status with the Unavailable code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewUnavailableStatus(message string) *status.Status {
	return status.New(codes.Unavailable, message)
}

// NewUnauthenticatedStatus returns a GRPC status with the Unauthenticated code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewUnauthenticatedStatus(message string) *status.Status {
	return status.New(codes.Unauthenticated, message)
}

// NewPermissionDeniedStatus returns a GRPC status with the PermissionDenied code.
// Refer to https://github.com/grpc/grpc-go/blob/master/codes/codes.go for the
// meaning of the error code.
func NewPermissionDeniedStatus(message string) *status.Status {
	return status.New(codes.PermissionDenied, message)
}
