package token_test

import (
	"context"
	"fmt"
	"p2pderivatives-server/internal/common/contexts"
	"p2pderivatives-server/internal/common/grpc/methods"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/test"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	validToken        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDg5MzYwMDAsImp0aSI6InV1aWQifQ.0muLv3oOrCU1Rj8IJvsYqcWd0bE-UWVnC9y8afxRJ0Q"
	expiredToken      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDYzNDU4MDAsImp0aSI6InRlc3QxIn0._HUTrOKAtYzLLUrMzhpA7TOrkl1NEp_M5YoRDDZsDmg"
	invalidToken      = "12345"
	emptyToken        = ""
	userID            = "uuid"
	noUserID          = "noUserID"
	success           = "success"
	noResult          = ""
	noTokenError      = "rpc error: code = InvalidArgument desc = accessToken required"
	invalidTokenError = "rpc error: code = FailedPrecondition desc = accessToken invalid"
	expiredTokenError = "rpc error: code = FailedPrecondition desc = accessToken expired"
	noError           = ""
)

func TestTokenInterceptor_WithNoTokenNotRequired_Succeed(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		unaryInterceptor,
		"TestNoToken",
		noUserID,
		success,
		noError)
}

func TestTokenInterceptor_WithTokenIsRequired_Succeed(t *testing.T) {
	testTokenHelper(
		newContext(validToken),
		t,
		unaryInterceptor,
		"TestWithToken",
		userID,
		success,
		noError)
}

func TestTokenInterceptor_WithNoTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		unaryInterceptor,
		"TestWithToken",
		noUserID,
		noResult,
		noTokenError)
}

func TestTokenInterceptor_WithExpiredTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		newContext(expiredToken),
		t,
		unaryInterceptor,
		"TestWithToken",
		noUserID,
		noResult,
		expiredTokenError,
	)
}

func TestTokenInterceptor_WithEmptyTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		newContext(emptyToken),
		t,
		unaryInterceptor,
		"TestWithToken",
		noUserID,
		noResult,
		noTokenError,
	)
}

func TestTokenInterceptor_WithInvalidTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		newContext(invalidToken),
		t,
		unaryInterceptor,
		"TestWithToken",
		noUserID,
		noResult,
		invalidTokenError)
}

func TestRequestStreamTokenInterceptor_WithNoTokenNotRequired_Success(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		streamInterceptor,
		"TestRequestStreamNoToken",
		noUserID,
		noResult,
		noError)
}

func TestRequestStreamTokenInterceptor_WithTokenIsRequired_Success(t *testing.T) {
	testTokenHelper(
		newContext(validToken),
		t,
		streamInterceptor,
		"TestRequestStreamWithToken",
		userID,
		noResult,
		noError)
}

func TestRequestStreamTokenInterceptor_WithNoTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		streamInterceptor,
		"TestRequestStreamWithToken",
		noUserID,
		noResult,
		noTokenError)
}

func TestRequestStreamTokenInterceptor_WithInvalidTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		newContext(invalidToken),
		t,
		streamInterceptor,
		"TestRequestStreamWithToken",
		noUserID,
		noResult,
		invalidTokenError)
}

func TestResponseStreamTokenInterceptor_WithNoTokenNotRequired_Success(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		streamInterceptor,
		"TestResponseStreamNoToken",
		noUserID,
		noResult,
		noError)
}

func TestResponseStreamTokenInterceptor_WithTokenIsRequired_Success(t *testing.T) {
	testTokenHelper(
		newContext(validToken),
		t,
		streamInterceptor,
		"TestResponseStreamWithToken",
		userID,
		noResult,
		noError)
}

func TestResponseStreamTokenInterceptor_WithNoTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		streamInterceptor,
		"TestResponseStreamWithToken",
		noUserID,
		noResult,
		noTokenError)
}

func TestResponseStreamTokenInterceptor_WithInvalidTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		newContext(invalidToken),
		t,
		streamInterceptor,
		"TestResponseStreamWithToken",
		noUserID,
		noResult,
		invalidTokenError)
}

func TestBiDirectionalStreamTokenInterceptor_WithNoTokenNotRequired_Success(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		streamInterceptor,
		"TestBiDirectionalStreamNoToken",
		noUserID,
		noResult,
		noError)
}

func TestBiDirectionalStreamTokenInterceptor_WithTokenIsRequired_Success(t *testing.T) {
	testTokenHelper(
		newContext(validToken),
		t,
		streamInterceptor,
		"TestBiDirectionalStreamWithToken",
		userID,
		noResult,
		noError)
}

func TestBiDirectionalStreamTokenInterceptor_WithNoTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		context.Background(),
		t,
		streamInterceptor,
		"TestBiDirectionalStreamWithToken",
		noUserID,
		noResult,
		noTokenError)
}

func TestBiDirectionalStreamTokenInterceptor_WithInvalidTokenIsRequired_Fails(t *testing.T) {
	testTokenHelper(
		newContext(invalidToken),
		t,
		streamInterceptor,
		"TestBiDirectionalStreamWithToken",
		noUserID,
		noResult,
		invalidTokenError)
}

func testTokenHelper(
	ctx context.Context,
	t *testing.T,
	interceptor func(context.Context, string, string, *assert.Assertions) (interface{}, error),
	methodName,
	expectedUserID,
	expectedResult,
	expectedError string) {
	// Arrange
	patch := monkey.Patch(time.Now, func() time.Time {
		return time.Date(2019, 1, 31, 12, 0, 0, 0, time.UTC)
	})
	defer patch.Unpatch()
	assert := assert.New(t)
	srv := grpc.NewServer()
	test.RegisterTestServer(srv, &test.Controller{})
	methods.Init(srv)
	token.Init(&token.Config{
		Secret:     "k^Cc#*mdnS9$nTOY6S1#1i7^e*o1ijSl",
		Exp:        time.Minute * 30,
		RefreshExp: time.Hour * 24 * 30,
	})

	// Act
	result, err := interceptor(ctx, methodName, expectedUserID, assert)

	// Assert
	if expectedError == noError {
		assert.NoError(err)
	} else {
		assert.EqualError(err, expectedError)
	}

	if expectedResult != noResult {
		assert.Equal(expectedResult, result)
	}
}

func getFullMethod(methodName string) string {
	return fmt.Sprintf("/test.Test/%s", methodName)
}

func unaryInterceptor(
	ctx context.Context, methodName, expectedUserID string,
	assert *assert.Assertions) (interface{}, error) {
	return token.UnaryInterceptor()(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: getFullMethod(methodName),
	}, unaryHandler(expectedUserID, assert))
}

func unaryHandler(
	expectedUserID string, assert *assert.Assertions) func(
	ctx context.Context, req interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		assertUserId(ctx, assert, expectedUserID)
		return success, nil
	}
}

func assertUserId(
	ctx context.Context, assert *assert.Assertions, expectedUserID string) {
	if expectedUserID == noUserID {
		assert.Panics(func() { contexts.GetUserID(ctx) })
	} else {
		assert.Equal(expectedUserID, contexts.GetUserID(ctx))
	}
}

func streamInterceptor(
	ctx context.Context,
	methodName, expectedUserID string,
	assert *assert.Assertions) (interface{}, error) {
	return noResult, token.StreamInterceptor()(nil, &mockStream{MockContext: ctx}, &grpc.StreamServerInfo{
		FullMethod: getFullMethod(methodName),
	}, streamHandler(expectedUserID, assert))
}

func streamHandler(expectedUserID string, assert *assert.Assertions) func(interface{}, grpc.ServerStream) error {
	return func(srv interface{}, stream grpc.ServerStream) error {
		assertUserId(stream.Context(), assert, expectedUserID)
		return nil
	}
}

func newContext(token string) context.Context {
	if len(token) == 0 {
		return metadata.NewIncomingContext(
			context.Background(), metadata.Pairs())
	}
	return metadata.NewIncomingContext(
		context.Background(), metadata.Pairs("authorization", token))
}

type mockStream struct {
	grpc.ServerStream
	MockContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *mockStream) Context() context.Context {
	return w.MockContext
}
