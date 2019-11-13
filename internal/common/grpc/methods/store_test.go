package methods_test

import (
	"testing"

	"p2pderivatives-server/internal/common/grpc/methods"
	"p2pderivatives-server/internal/common/grpc/pbbase"
	"p2pderivatives-server/test"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMethodsInit_HasCorrectTokenMethods(t *testing.T) {
	srv := grpc.NewServer()
	test.RegisterTestServer(srv, &test.Controller{})
	methods.Init(srv)

	assert.True(t, methods.IsIgnoreTokenVerify("/test.Test/TestNoToken"))
	assert.True(t, methods.IsIgnoreTokenVerify("/test.Test/TestRequestStreamNoToken"))
	assert.True(t, methods.IsIgnoreTokenVerify("/test.Test/TestResponseStreamNoToken"))
	assert.True(t, methods.IsIgnoreTokenVerify("/test.Test/TestBiDirectionalStreamNoToken"))
	assert.False(t, methods.IsIgnoreTokenVerify("/test.Test/TestWithToken"))
}

func TestMethodsInit_HasCorrectTxOptionsMethods(t *testing.T) {
	srv := grpc.NewServer()
	test.RegisterTestServer(srv, &test.Controller{})
	methods.Init(srv)

	assert.Equal(t, pbbase.TxOption_ReadWrite, methods.TxOption("/test.Test/TestDefaultTxOption"))
	assert.Equal(t, pbbase.TxOption_ReadWrite, methods.TxOption("/test.Test/TestReadWriteTxOption"))
	assert.Equal(t, pbbase.TxOption_ReadOnly, methods.TxOption("/test.Test/TestReadOnlyTxOption"))
	assert.Equal(t, pbbase.TxOption_NoTx, methods.TxOption("/test.Test/TestNoTxOption"))
}
