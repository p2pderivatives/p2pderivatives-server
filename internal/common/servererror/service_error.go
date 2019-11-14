package servererror

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/pkg/errors"
)

// ServiceError contain a log entry, and is used to create and log an error from
// a service.
type ServiceError struct {
}

// CreateServiceError creates and log an error with the given information.
func (s *ServiceError) CreateServiceError(ctx context.Context, code ErrorCode, message string, err error) error {
	log := ctxlogrus.Extract(ctx)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(fmt.Sprintf("%s error:%+v", message, err))
	} else {
		log.Errorf(message)
	}
	return NewError(code, message, err)
}

// CreateServiceErrorWithDetail creates and logs a detailed error with the given
// information.
func (s *ServiceError) CreateServiceErrorWithDetail(
	ctx context.Context,
	code ErrorCode,
	message string,
	err error,
	detailCode ErrorDetailCode,
	detailValues []string) error {
	log := ctxlogrus.Extract(ctx)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(fmt.Sprintf("%s error:%+v", message, err))
	} else {
		log.Errorf(message)
	}
	return NewErrorWithDetail(code, message, err, detailCode, detailValues)
}

// CreateServiceErrorWithDetails creates and logs a detailed error with the given
// information.
func (s *ServiceError) CreateServiceErrorWithDetails(
	ctx context.Context,
	code ErrorCode,
	message string,
	err error,
	errDetails []ErrorDetail) error {
	log := ctxlogrus.Extract(ctx)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(fmt.Sprintf("%s error:%+v", message, err))
	} else {
		log.Errorf(message)
	}
	return NewErrorWithDetails(code, message, err, errDetails)
}
