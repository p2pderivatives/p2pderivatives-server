package token

import (
	"context"
	"p2pderivatives-server/internal/common/contexts"
	"p2pderivatives-server/internal/common/grpc/methods"
	"p2pderivatives-server/internal/common/servererror"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	//MetaKeyAuthentication is the key for setting and retrieving from metadata.
	MetaKeyAuthentication = "authorization"
)

var (
	// ErrInvalidRequest is an error returned when a token was expected but
	// not provided.
	ErrInvalidRequest = servererror.NewErrorWithDetail(servererror.InvalidArguments, "accessToken required", nil, servererror.ErrorDetailCodeTokenRequired, nil)
	// ErrTokenExpired is an error returned when a token was expected but the
	// provided one was expired.
	ErrTokenExpired = servererror.NewErrorWithDetail(servererror.PreconditionError, "accessToken expired", nil, servererror.ErrorDetailCodeTokenExpired, nil)
	// ErrTokenInvalid is an error returned when a token was expected but the
	// provided one was invalid.
	ErrTokenInvalid = servererror.NewErrorWithDetail(servererror.PreconditionError, "accessToken invalid", nil, servererror.ErrorDetailCodeTokenInvalid, nil)
)

// UnaryInterceptor is a unary interceptor that checks that a valid token is
// included in the metadata of the request.
func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if methods.IsIgnoreTokenVerify(info.FullMethod) {
			return handler(ctx, req)
		}

		newCtx, err := verifyToken(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// StreamInterceptor is a stream interceptor that checks that a valid token is
// provided in the metadata of the request.
func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if methods.IsIgnoreTokenVerify(info.FullMethod) {
			return handler(srv, stream)
		}
		newCtx, err := verifyToken(stream.Context())
		if err != nil {
			return err
		}
		return handler(srv, &wrappedStream{stream, newCtx})
	}
}

type wrappedStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested
// grpc.ServerStream.Context()
func (w *wrappedStream) Context() context.Context {
	return w.WrappedContext
}

func verifyToken(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, servererror.GetGrpcStatus(ctx, ErrInvalidRequest).Err()
	}
	vals := md.Get(MetaKeyAuthentication)
	if len(vals) == 0 {
		return ctx, servererror.GetGrpcStatus(ctx, ErrInvalidRequest).Err()
	}
	accessToken := vals[0]
	id, err := VerifyToken(accessToken)
	if err != nil {
		if IsTokenExpiredError(err) {
			return ctx, servererror.GetGrpcStatus(ctx, err).Err()
		}
		return ctx, servererror.GetGrpcStatus(ctx, ErrTokenInvalid).Err()
	}
	return contexts.SetUserID(ctx, id), nil
}
