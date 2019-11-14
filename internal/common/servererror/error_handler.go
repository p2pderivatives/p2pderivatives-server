package servererror

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GetGrpcStatus converts the given error to a GRPC status corresponding to the
// code contained in the error. If the error is not a service.Error instance,
// it will be transformed to an UnknownError and an InternalStatus will be
// returned.
func GetGrpcStatus(ctx context.Context, err error) *status.Status {
	serr, ok := err.(*Error)
	if !ok {
		serr = NewError(UnknownError, "An undefined error occurred.", err).(*Error)
	}

	if len(serr.Details) != 0 {
		bytes, err := json.Marshal(serr.Details)
		var base64json string
		if err != nil {
			base64json = string(base64.StdEncoding.EncodeToString([]byte("Failed to convert detail")))
		} else {
			base64json = string(base64.StdEncoding.EncodeToString(bytes))
		}
		trailer := metadata.Pairs("x-error-detail", base64json)
		grpc.SetTrailer(ctx, trailer)
	}

	switch serr.Code {
	case InternalError:
		return NewInternalStatus(serr.Message)
	case DbError:
		return NewInternalStatus(serr.Message)
	case InvalidArguments:
		return NewInvalidArgumentStatus(serr.Message)
	case DeadlineExceeded:
		return NewDeadlineExceededStatus(serr.Message)
	case NotFoundError:
		return NewNotFoundStatus(serr.Message)
	case AlreadyExistError:
		return NewAlreadyExistsStatus(serr.Message)
	case OptimisticLockError:
		return NewFailedPreconditionStatus(serr.Message)
	case PreconditionError:
		return NewFailedPreconditionStatus(serr.Message)
	case Unavailable:
		return NewUnavailableStatus(serr.Message)
	case UnauthenticatedError:
		return NewUnauthenticatedStatus(serr.Message)
	case UnknownError:
		return NewInternalStatus(serr.Message)
	case PermissionDenied:
		return NewPermissionDeniedStatus(serr.Message)
	default:
		return NewInternalStatus(serr.Message)
	}
}
