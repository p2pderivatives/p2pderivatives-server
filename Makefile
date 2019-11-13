gen-proto:
	#pbbase/*.proto
	mkdir -p ./internal/common/grpc/pbbase
	protoc -I/usr/local/include -I./api/proto -I${GOPATH}/src \
		--go_out=./internal/common/grpc/pbbase\
		./api/proto/method_option.proto
	#test/*.proto
	protoc -I./test -I./api/proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./test test.proto

