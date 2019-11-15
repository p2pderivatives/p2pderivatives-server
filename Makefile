gen:
	@make gen-proto
	@make gen-mock

gen-proto:
	#pbbase/*.proto
	mkdir -p ./internal/common/grpc/pbbase
	protoc -I/usr/local/include -I./api/proto -I${GOPATH}/src \
		--go_out=./internal/common/grpc/pbbase\
		./api/proto/method_option.proto
	#user/*.proto
	protoc -I./internal/user -I./api/proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./internal/user/usercontroller user.proto \
		--govalidators_out=gogoimport=true:./internal/user/usercontroller user.proto
	#authentication/*.proto
	protoc -I./internal/authentication -I./api/proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./internal/authentication authentication.proto \
		--govalidators_out=gogoimport=true:./internal/authentication authentication.proto
	#test/*.proto
	protoc -I./test -I./api/proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./test test.proto

gen-mock:
	mockgen -destination test/mocks/mock_usercontroller/mock_controller.go  p2pderivatives-server/internal/user/usercontroller User_GetUserStatusesServer,User_GetUserListServer
	mockgen -destination test/mocks/mock_usercommon/mock_service.go  p2pderivatives-server/internal/user/usercommon ServiceIf
