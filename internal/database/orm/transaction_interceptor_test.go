package orm_test

import (
	"context"
	"testing"

	"p2pderivatives-server/internal/common/grpc/pbbase"
	"p2pderivatives-server/internal/database/orm"
	"p2pderivatives-server/test"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestTransactionInterceptorUnaryInterceptor_RWOption_HasTx(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	var retrievedTx *gorm.DB

	// Act
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		retrievedTx = orm.ExtractTx(ctx)
		// TODO(tibo): would be nice to be able to check that it's RW.
		return nil, nil
	}

	unaryInterceptorTestHelper(pbbase.TxOption_ReadWrite, handler)

	// Assert
	assert.NotNil(retrievedTx)
}

func TestTransactionInterceptorUnaryInterceptor_ReadOnlyTxOption_HasTx(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	var retrievedTx *gorm.DB

	// Act
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		retrievedTx = orm.ExtractTx(ctx)
		// TODO(tibo): would be nice to be able to check that it's RW.
		return nil, nil
	}

	unaryInterceptorTestHelper(pbbase.TxOption_ReadOnly, handler)

	// Assert
	assert.NotNil(retrievedTx)
}

func TestTransactionInterceptorUnaryInterceptor_NoTxOption_Panics(t *testing.T) {
	// Arrange
	assert := assert.New(t)

	// Act/Assert
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.Panics(func() { orm.ExtractTx(ctx) })
		return nil, nil
	}

	unaryInterceptorTestHelper(pbbase.TxOption_NoTx, handler)
}

func TestTransactionInterceptorStreamInterceptor_RWTxOption_HasTx(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	var retrievedTx *gorm.DB

	// Act
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		retrievedTx = orm.ExtractTx(stream.Context())
		return nil
	}

	streamInterceptorTestHelper(pbbase.TxOption_ReadWrite, handler)

	// Assert
	assert.NotNil(retrievedTx)
}

func TestTransactionInterceptorStreamInterceptor_ReadOnlyTxOption_HasTx(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	var retrievedTx *gorm.DB

	// Act
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		retrievedTx = orm.ExtractTx(stream.Context())
		return nil
	}

	streamInterceptorTestHelper(pbbase.TxOption_ReadOnly, handler)

	// Assert
	assert.NotNil(retrievedTx)
}

func TestTransactionInterceptorStreamInterceptor_NoTxOption_Panics(t *testing.T) {
	// Arrange
	assert := assert.New(t)

	// Act/Assert
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		assert.Panics(func() { orm.ExtractTx(stream.Context()) })
		return nil
	}

	streamInterceptorTestHelper(pbbase.TxOption_NoTx, handler)
}

func unaryInterceptorTestHelper(
	txOption pbbase.TxOption,
	handler func(context.Context, interface{}) (interface{}, error)) {
	// Arrange
	config := test.GetTestConfig()
	log := test.GetTestLogger(config)
	ormInstance := test.InitializeORM()
	defer ormInstance.Finalize()
	tx := ormInstance.GetDB().Begin()
	defer tx.Rollback()

	// Act
	orm.TransactionUnaryServerInterceptor(
		log.NewEntry(),
		func(s string) pbbase.TxOption { return txOption },
		ormInstance)(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{},
		handler)

}

func streamInterceptorTestHelper(
	txOption pbbase.TxOption,
	handler func(srv interface{}, stream grpc.ServerStream) error) {
	// Arrange
	config := test.GetTestConfig()
	log := test.GetTestLogger(config)
	ormInstance := test.InitializeORM()
	defer ormInstance.Finalize()
	tx := ormInstance.GetDB().Begin()
	defer tx.Rollback()

	// Act
	orm.TransactionStreamServerInterceptor(
		log.NewEntry(),
		func(s string) pbbase.TxOption { return txOption },
		ormInstance)(
		nil,
		&mockStream{MockContext: context.Background()},
		&grpc.StreamServerInfo{},
		handler)

}

type mockStream struct {
	grpc.ServerStream
	MockContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *mockStream) Context() context.Context {
	return w.MockContext
}
