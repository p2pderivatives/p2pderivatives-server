package interceptor

import (
	"context"
	"database/sql"
	"p2pderivatives-server/internal/common/grpc/pbbase"

	"github.com/cryptogarageinc/server-common-go/pkg/database/orm"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ctxTxMarker is used to retrieve the DB transaction from the context.
type ctxTxMarker struct{}

var (
	ctxTxKey = &ctxTxMarker{}
)

type wrappedStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedStream) Context() context.Context {
	return w.WrappedContext
}

// TransactionStreamServerInterceptor provides the DB transaction for streaming
// methods which require it, and rollbacks in case of errors.
func TransactionStreamServerInterceptor(
	log *logrus.Entry,
	txOption func(fullMethod string) pbbase.TxOption,
	ormInstance *orm.ORM) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		subHandler := func(ctx context.Context) (interface{}, error) {
			wStream := &wrappedStream{ss, ctx}
			err := handler(srv, wStream)
			return nil, err
		}

		_, err := interceptor(
			ss.Context(), log, txOption, ormInstance, info.FullMethod, subHandler)
		return err
	}
}

// TransactionUnaryServerInterceptor provides the DB transaction for methods
// which require it, and rollbacks in case of errors.
func TransactionUnaryServerInterceptor(
	log *logrus.Entry,
	txOption func(fullMethod string) pbbase.TxOption,
	ormInstance *orm.ORM) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		subHandler := func(ctx context.Context) (interface{}, error) {
			res, err := handler(ctx, req)
			return res, err
		}

		res, err := interceptor(
			ctx, log, txOption, ormInstance, info.FullMethod, subHandler)
		return res, err
	}
}

// SaveTx adds the DB transaction to the context.
func SaveTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxTxKey, tx)
}

// ExtractTx extracts the DB transaction from the context.
func ExtractTx(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(ctxTxKey).(*gorm.DB)
	if !ok || tx == nil {
		panic("Could not retrieve transaction from context.")
	}
	return tx
}

func interceptor(
	ctx context.Context,
	log *logrus.Entry,
	txOption func(methodName string) pbbase.TxOption,
	ormInstance *orm.ORM,
	fullMethod string,
	handler func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	log = log.WithField("method", fullMethod)

	switch txOption(fullMethod) {
	case pbbase.TxOption_NoTx:
		return handler(ctx)
	case pbbase.TxOption_ReadOnly:
		return handleReadOnly(ctx, log, ormInstance, handler)
	case pbbase.TxOption_ReadWrite:
		return handleReadWrite(ctx, log, ormInstance, handler)
	}

	panic("Unhandled tx option.")
}

func handleReadOnly(
	ctx context.Context,
	log *logrus.Entry,
	ormInstance *orm.ORM,
	handler func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	options := &sql.TxOptions{ReadOnly: true}
	tx := ormInstance.GetDB().Begin(options)
	res, err := handler(SaveTx(ctx, tx))
	tx.Rollback()
	return res, err
}

func handleReadWrite(
	ctx context.Context,
	log *logrus.Entry,
	ormInstance *orm.ORM,
	handler func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	tx := ormInstance.GetDB().Begin()
	newCtx := SaveTx(ctx, tx)

	res, err := handler(newCtx)

	if err != nil {
		if tx.Rollback(); tx.Error != nil {
			log.Errorf("failed to rollback: %+v", tx.Error)
		}
		log.Infoln("DB transaction for the method handler is rollbacked", err)
		return nil, err
	}

	if tx.Commit(); tx.Error != nil {
		log.Errorf("failed to commit: %+v", tx.Error)
		return nil, status.Errorf(codes.Internal, "failed to commit")
	}
	log.Debugf("commit succeeds")
	return res, nil
}
