package test

import context "context"

// Controller is respresent a controller used for testing JWT tokens
type Controller struct{}

// TestNoToken basic test function not requiring token.
func (controller *Controller) TestNoToken(
	ctx context.Context, empty *Empty) (*Response, error) {
	return &Response{Ok: true}, nil
}

// TestWithToken basic test function requiring token.
func (controller *Controller) TestWithToken(
	ctx context.Context, empty *Empty) (*Response, error) {
	return &Response{Ok: true}, nil
}

// TestRequestStreamNoToken request stream test function not requiring token.
func (controller *Controller) TestRequestStreamNoToken(
	stream Test_TestRequestStreamNoTokenServer) error {
	return stream.SendAndClose(&Response{Ok: true})
}

// TestRequestStreamWithToken request stream test function requiring token.
func (controller *Controller) TestRequestStreamWithToken(
	stream Test_TestRequestStreamWithTokenServer) error {
	return stream.SendAndClose(&Response{Ok: true})
}

// TestResponseStreamNoToken response stream test function not requiring token.
func (controller *Controller) TestResponseStreamNoToken(
	empty *Empty, stream Test_TestResponseStreamNoTokenServer) error {
	stream.Send(&Response{Ok: true})
	return nil
}

// TestResponseStreamWithToken response stream test function requiring token.
func (controller *Controller) TestResponseStreamWithToken(
	empty *Empty, stream Test_TestResponseStreamWithTokenServer) error {
	stream.Send(&Response{Ok: true})
	return nil
}

// TestBiDirectionalStreamNoToken bi-directional stream test function not
// requiring token.
func (controller *Controller) TestBiDirectionalStreamNoToken(
	stream Test_TestBiDirectionalStreamNoTokenServer) error {
	stream.Send(&Response{Ok: true})
	return nil
}

// TestBiDirectionalStreamWithToken bi-directional stream test function
// requiring token.
func (controller *Controller) TestBiDirectionalStreamWithToken(
	stream Test_TestBiDirectionalStreamWithTokenServer) error {
	stream.Send(&Response{Ok: true})
	return nil
}

// TestDefaultTxOption method with no specified tx option.
func (controller *Controller) TestDefaultTxOption(
	ctx context.Context, empty *Empty) (*Empty, error) {
	return &Empty{}, nil
}

// TestReadWriteTxOption method with read/write tx option.
func (controller *Controller) TestReadWriteTxOption(
	ctx context.Context, empty *Empty) (*Empty, error) {
	return &Empty{}, nil
}

// TestReadOnlyTxOption method with read-only tx option.
func (controller *Controller) TestReadOnlyTxOption(
	ctx context.Context, empty *Empty) (*Empty, error) {
	return &Empty{}, nil
}

// TestNoTxOption method with no tx tx option.
func (controller *Controller) TestNoTxOption(
	ctx context.Context, empty *Empty) (*Empty, error) {
	return &Empty{}, nil
}
